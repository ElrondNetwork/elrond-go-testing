package smartContract

import (
	"bytes"
	"encoding/hex"
	"math/big"
	"sync"

	"github.com/ElrondNetwork/elrond-go-testing/core"
	"github.com/ElrondNetwork/elrond-go-testing/core/check"
	"github.com/ElrondNetwork/elrond-go-testing/data"
	"github.com/ElrondNetwork/elrond-go-testing/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go-testing/data/state"
	"github.com/ElrondNetwork/elrond-go-testing/data/transaction"
	"github.com/ElrondNetwork/elrond-go-testing/hashing"
	"github.com/ElrondNetwork/elrond-go-testing/logger"
	"github.com/ElrondNetwork/elrond-go-testing/marshal"
	"github.com/ElrondNetwork/elrond-go-testing/process"
	"github.com/ElrondNetwork/elrond-go-testing/sharding"
	"github.com/ElrondNetwork/elrond-vm-common"
)

var log = logger.GetOrCreate("process/smartcontract")

type scExecutionState struct {
	allLogs       map[string][]*vmcommon.LogEntry
	allReturnData map[string][]*big.Int
	returnCodes   map[string]vmcommon.ReturnCode
	rootHash      []byte
}

type scProcessor struct {
	accounts         state.AccountsAdapter
	tempAccounts     process.TemporaryAccountsHandler
	adrConv          state.AddressConverter
	hasher           hashing.Hasher
	marshalizer      marshal.Marshalizer
	shardCoordinator sharding.Coordinator
	vmContainer      process.VirtualMachinesContainer
	argsParser       process.ArgumentsParser

	mutSCState   sync.Mutex
	mapExecState map[uint64]scExecutionState

	scrForwarder process.IntermediateTransactionHandler
	txFeeHandler process.TransactionFeeHandler
	economicsFee process.FeeHandler
	gasHandler   process.GasHandler
}

// NewSmartContractProcessor create a smart contract processor creates and interprets VM data
func NewSmartContractProcessor(
	vmContainer process.VirtualMachinesContainer,
	argsParser process.ArgumentsParser,
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	accountsDB state.AccountsAdapter,
	tempAccounts process.TemporaryAccountsHandler,
	adrConv state.AddressConverter,
	coordinator sharding.Coordinator,
	scrForwarder process.IntermediateTransactionHandler,
	txFeeHandler process.TransactionFeeHandler,
	economicsFee process.FeeHandler,
	gasHandler process.GasHandler,
) (*scProcessor, error) {
	err := checkArgumentsForNil(vmContainer, argsParser, hasher, marshalizer, accountsDB,
		tempAccounts, adrConv, coordinator, scrForwarder, txFeeHandler, economicsFee, gasHandler)
	if err != nil {
		return nil, err
	}

	return &scProcessor{
		vmContainer:      vmContainer,
		argsParser:       argsParser,
		hasher:           hasher,
		marshalizer:      marshalizer,
		accounts:         accountsDB,
		tempAccounts:     tempAccounts,
		adrConv:          adrConv,
		shardCoordinator: coordinator,
		scrForwarder:     scrForwarder,
		txFeeHandler:     txFeeHandler,
		economicsFee:     economicsFee,
		gasHandler:       gasHandler,
		mapExecState:     make(map[uint64]scExecutionState)}, nil
}

func checkArgumentsForNil(
	vmContainer process.VirtualMachinesContainer,
	argsParser process.ArgumentsParser,
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	accountsDB state.AccountsAdapter,
	tempAccounts process.TemporaryAccountsHandler,
	adrConv state.AddressConverter,
	coordinator sharding.Coordinator,
	scrForwarder process.IntermediateTransactionHandler,
	txFeeHandler process.TransactionFeeHandler,
	economicsFee process.FeeHandler,
	gasHandler process.GasHandler,
) error {
	if check.IfNil(vmContainer) {
		return process.ErrNoVM
	}
	if check.IfNil(argsParser) {
		return process.ErrNilArgumentParser
	}
	if check.IfNil(hasher) {
		return process.ErrNilHasher
	}
	if check.IfNil(marshalizer) {
		return process.ErrNilMarshalizer
	}
	if check.IfNil(accountsDB) {
		return process.ErrNilAccountsAdapter
	}
	if check.IfNil(tempAccounts) {
		return process.ErrNilTemporaryAccountsHandler
	}
	if check.IfNil(adrConv) {
		return process.ErrNilAddressConverter
	}
	if check.IfNil(coordinator) {
		return process.ErrNilShardCoordinator
	}
	if check.IfNil(scrForwarder) {
		return process.ErrNilIntermediateTransactionHandler
	}
	if check.IfNil(txFeeHandler) {
		return process.ErrNilUnsignedTxHandler
	}
	if check.IfNil(economicsFee) {
		return process.ErrNilEconomicsFeeHandler
	}
	if check.IfNil(gasHandler) {
		return process.ErrNilGasHandler
	}

	return nil
}

func (sc *scProcessor) checkTxValidity(tx *transaction.Transaction) error {
	if tx == nil || tx.IsInterfaceNil() {
		return process.ErrNilTransaction
	}

	recvAddressIsInvalid := sc.adrConv.AddressLen() != len(tx.RcvAddr)
	if recvAddressIsInvalid {
		return process.ErrWrongTransaction
	}

	return nil
}

func (sc *scProcessor) isDestAddressEmpty(tx *transaction.Transaction) bool {
	isEmptyAddress := bytes.Equal(tx.RcvAddr, make([]byte, sc.adrConv.AddressLen()))
	return isEmptyAddress
}

func (sc *scProcessor) computeTransactionHash(tx data.TransactionHandler) ([]byte, error) {
	scr, ok := tx.(*smartContractResult.SmartContractResult)
	if ok {
		return scr.TxHash, nil
	}

	return core.CalculateHash(sc.marshalizer, sc.hasher, tx)
}

func (sc *scProcessor) createSCRsWhenError(
	tx data.TransactionHandler,
	returnCode vmcommon.ReturnCode,
) ([]data.TransactionHandler, error) {
	txHash, err := sc.computeTransactionHash(tx)
	if err != nil {
		return nil, err
	}

	scr := &smartContractResult.SmartContractResult{
		Nonce:   tx.GetNonce(),
		Value:   tx.GetValue(),
		RcvAddr: tx.GetSndAddress(),
		SndAddr: tx.GetRecvAddress(),
		Code:    nil,
		Data:    "@" + hex.EncodeToString([]byte(returnCode.String())) + "@" + hex.EncodeToString(txHash),
		TxHash:  txHash,
	}

	resultedScrs := make([]data.TransactionHandler, 0)
	resultedScrs = append(resultedScrs, scr)

	return resultedScrs, nil
}

// ExecuteSmartContractTransaction processes the transaction, call the VM and processes the SC call output
func (sc *scProcessor) ExecuteSmartContractTransaction(
	tx *transaction.Transaction,
	acntSnd, acntDst state.AccountHandler,
	round uint64,
) error {
	defer sc.tempAccounts.CleanTempAccounts()

	if tx == nil || tx.IsInterfaceNil() {
		return process.ErrNilTransaction
	}
	if acntDst == nil || acntDst.IsInterfaceNil() {
		return process.ErrNilSCDestAccount
	}
	if acntDst.IsInterfaceNil() || acntDst.GetCode() == nil {
		return process.ErrNilSCDestAccount
	}

	err := sc.processSCPayment(tx, acntSnd)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			err = sc.processIfError(acntSnd, tx, vmcommon.UserError)
			if err != nil {
				log.Debug("error while processing error in smart contract processor")
			}
		}
	}()

	err = sc.prepareSmartContractCall(tx, acntSnd)
	if err != nil {
		return nil
	}

	vmInput, err := sc.createVMCallInput(tx)
	if err != nil {
		return nil
	}

	vm, err := sc.getVMFromRecvAddress(tx)
	if err != nil {
		return nil
	}

	vmOutput, err := vm.RunSmartContractCall(vmInput)
	if err != nil {
		return nil
	}

	results, consumedFee, err := sc.processVMOutput(vmOutput, tx, acntSnd, round)
	if err != nil {
		return nil
	}

	err = sc.scrForwarder.AddIntermediateTransactions(results)
	if err != nil {
		return nil
	}

	sc.txFeeHandler.ProcessTransactionFee(consumedFee)

	return nil
}

func (sc *scProcessor) processIfError(
	acntSnd state.AccountHandler,
	tx data.TransactionHandler,
	returnCode vmcommon.ReturnCode,
) error {
	consumedFee := big.NewInt(0).SetUint64(tx.GetGasLimit() * tx.GetGasPrice())
	scrIfError, err := sc.createSCRsWhenError(tx, returnCode)
	if err != nil {
		return err
	}

	if check.IfNil(acntSnd) {
		return nil
	}

	stAcc, ok := acntSnd.(*state.Account)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	totalCost := big.NewInt(0)
	err = stAcc.SetBalanceWithJournal(totalCost.Add(stAcc.Balance, tx.GetValue()))
	if err != nil {
		return err
	}

	err = sc.scrForwarder.AddIntermediateTransactions(scrIfError)
	if err != nil {
		return nil
	}

	sc.txFeeHandler.ProcessTransactionFee(consumedFee)

	return nil
}

func (sc *scProcessor) prepareSmartContractCall(tx *transaction.Transaction, acntSnd state.AccountHandler) error {
	err := sc.argsParser.ParseData(tx.Data)
	if err != nil {
		return err
	}

	nonce := tx.Nonce
	if acntSnd != nil && !acntSnd.IsInterfaceNil() {
		nonce = acntSnd.GetNonce()
	}

	txValue := big.NewInt(0).Set(tx.Value)
	sc.tempAccounts.AddTempAccount(tx.SndAddr, txValue, nonce)

	return nil
}

func (sc *scProcessor) getVMTypeFromArguments(vmType []byte) ([]byte, error) {
	// first parsed argument after the code in case of vmDeploy is the actual vmType
	vmAppendedType := make([]byte, core.VMTypeLen)
	vmArgLen := len(vmType)
	if vmArgLen > core.VMTypeLen {
		return nil, process.ErrVMTypeLengthInvalid
	}

	copy(vmAppendedType[core.VMTypeLen-vmArgLen:], vmType)
	return vmAppendedType, nil
}

func (sc *scProcessor) getVMFromRecvAddress(tx *transaction.Transaction) (vmcommon.VMExecutionHandler, error) {
	vmType := core.GetVMType(tx.RcvAddr)
	vm, err := sc.vmContainer.Get(vmType)
	if err != nil {
		return nil, err
	}
	return vm, nil
}

// DeploySmartContract processes the transaction, than deploy the smart contract into VM, final code is saved in account
func (sc *scProcessor) DeploySmartContract(
	tx *transaction.Transaction,
	acntSnd state.AccountHandler,
	round uint64,
) error {
	defer sc.tempAccounts.CleanTempAccounts()

	err := sc.checkTxValidity(tx)
	if err != nil {
		log.Debug("Transaction invalid", "error", err.Error())
		return err
	}

	isEmptyAddress := sc.isDestAddressEmpty(tx)
	if !isEmptyAddress {
		log.Debug("Transaction wrong", "error", process.ErrWrongTransaction.Error())
		return process.ErrWrongTransaction
	}

	err = sc.processSCPayment(tx, acntSnd)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			err = sc.processIfError(acntSnd, tx, vmcommon.UserError)
			if err != nil {
				log.Debug("error while processing error in smart contract processor")
			}
		}
	}()

	err = sc.prepareSmartContractCall(tx, acntSnd)
	if err != nil {
		log.Debug("Transaction error", "error", err.Error())
		return nil
	}

	vmInput, vmType, err := sc.createVMDeployInput(tx)
	if err != nil {
		log.Debug("Transaction error", "error", err.Error())
		return nil
	}

	vm, err := sc.vmContainer.Get(vmType)
	if err != nil {
		log.Debug("VM error", "error", err.Error())
		return nil
	}

	vmOutput, err := vm.RunSmartContractCreate(vmInput)
	if err != nil {
		log.Debug("VM error", "error", err.Error())
		return nil
	}

	results, consumedFee, err := sc.processVMOutput(vmOutput, tx, acntSnd, round)
	if err != nil {
		log.Debug("Processing error", "error", err.Error())
		return nil
	}

	err = sc.scrForwarder.AddIntermediateTransactions(results)
	if err != nil {
		log.Debug("Processing error", "error", err.Error())
		return nil
	}

	sc.txFeeHandler.ProcessTransactionFee(consumedFee)

	log.Trace("SmartContract deployed")
	return nil
}

func (sc *scProcessor) createVMCallInput(tx *transaction.Transaction) (*vmcommon.ContractCallInput, error) {
	vmInput, err := sc.createVMInput(tx)
	if err != nil {
		return nil, err
	}

	vmCallInput := &vmcommon.ContractCallInput{}
	vmCallInput.VMInput = *vmInput
	vmCallInput.Function, err = sc.argsParser.GetFunction()
	if err != nil {
		return nil, err
	}

	vmCallInput.RecipientAddr = tx.RcvAddr

	return vmCallInput, nil
}

func (sc *scProcessor) createVMDeployInput(
	tx *transaction.Transaction,
) (*vmcommon.ContractCreateInput, []byte, error) {
	vmInput, err := sc.createVMInput(tx)
	if err != nil {
		return nil, nil, err
	}

	if len(vmInput.Arguments) < 1 {
		return nil, nil, process.ErrNotEnoughArgumentsToDeploy
	}

	vmType, err := sc.getVMTypeFromArguments(vmInput.Arguments[0])
	if err != nil {
		return nil, nil, err
	}
	// delete the first argument as it is the vmType
	vmInput.Arguments = vmInput.Arguments[1:]

	vmCreateInput := &vmcommon.ContractCreateInput{}
	hexCode, err := sc.argsParser.GetCode()
	if err != nil {
		return nil, nil, err
	}

	vmCreateInput.ContractCode, err = hex.DecodeString(string(hexCode))
	if err != nil {
		return nil, nil, err
	}

	vmCreateInput.VMInput = *vmInput

	return vmCreateInput, vmType, nil
}

func (sc *scProcessor) createVMInput(tx *transaction.Transaction) (*vmcommon.VMInput, error) {
	var err error
	vmInput := &vmcommon.VMInput{}

	vmInput.CallerAddr = tx.SndAddr
	vmInput.Arguments, err = sc.argsParser.GetArguments()
	if err != nil {
		return nil, err
	}
	vmInput.CallValue = tx.Value
	vmInput.GasPrice = tx.GasPrice
	moveBalanceGasConsume := sc.economicsFee.ComputeGasLimit(tx)
	vmInput.GasProvided = tx.GasLimit - moveBalanceGasConsume
	if tx.GetGasLimit() < moveBalanceGasConsume {
		return nil, process.ErrNotEnoughGas
	}

	return vmInput, nil
}

// taking money from sender, as VM might not have access to him because of state sharding
func (sc *scProcessor) processSCPayment(tx *transaction.Transaction, acntSnd state.AccountHandler) error {
	if acntSnd == nil || acntSnd.IsInterfaceNil() {
		// transaction was already processed at sender shard
		return nil
	}

	err := acntSnd.SetNonceWithJournal(acntSnd.GetNonce() + 1)
	if err != nil {
		return err
	}

	err = sc.economicsFee.CheckValidityTxValues(tx)
	if err != nil {
		return err
	}

	cost := big.NewInt(0)
	cost = cost.Mul(big.NewInt(0).SetUint64(tx.GasPrice), big.NewInt(0).SetUint64(tx.GasLimit))
	cost = cost.Add(cost, tx.Value)

	if cost.Cmp(big.NewInt(0)) == 0 {
		return nil
	}

	stAcc, ok := acntSnd.(*state.Account)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	if stAcc.Balance.Cmp(cost) < 0 {
		return process.ErrInsufficientFunds
	}

	totalCost := big.NewInt(0)
	err = stAcc.SetBalanceWithJournal(totalCost.Sub(stAcc.Balance, cost))
	if err != nil {
		return err
	}

	return nil
}

func (sc *scProcessor) processVMOutput(
	vmOutput *vmcommon.VMOutput,
	tx *transaction.Transaction,
	acntSnd state.AccountHandler,
	round uint64,
) ([]data.TransactionHandler, *big.Int, error) {
	if vmOutput == nil {
		return nil, nil, process.ErrNilVMOutput
	}
	if tx == nil {
		return nil, nil, process.ErrNilTransaction
	}

	txBytes, err := sc.marshalizer.Marshal(tx)
	if err != nil {
		return nil, nil, err
	}
	txHash := sc.hasher.Compute(string(txBytes))

	err = sc.saveSCOutputToCurrentState(vmOutput, round, txHash)
	if err != nil {
		return nil, nil, err
	}

	if vmOutput.ReturnCode != vmcommon.Ok {
		log.Debug("error processing tx VM",
			"hash", txHash,
			"return code", vmOutput.ReturnCode.String(),
		)

		err := sc.processIfError(acntSnd, tx, vmOutput.ReturnCode)
		return nil, nil, err
	}

	err = sc.processSCOutputAccounts(vmOutput.OutputAccounts, tx)
	if err != nil {
		return nil, nil, err
	}

	scrTxs, err := sc.createSCRTransactions(vmOutput.OutputAccounts, tx, txHash)
	if err != nil {
		return nil, nil, err
	}

	acntSnd, err = sc.reloadLocalSndAccount(acntSnd)
	if err != nil {
		return nil, nil, err
	}

	totalGasConsumed := tx.GasLimit - vmOutput.GasRemaining
	log.Debug("total gas consumed", "value", totalGasConsumed, "hash", txHash)

	if vmOutput.GasRefund.Uint64() > 0 {
		log.Debug("total gas refunded", "value", vmOutput.GasRefund.Uint64(), "hash", txHash)
	}

	totalGasRefund := big.NewInt(0)
	totalGasRefund = totalGasRefund.Add(vmOutput.GasRefund, big.NewInt(0).SetUint64(vmOutput.GasRemaining))
	sc.gasHandler.SetGasRefunded(vmOutput.GasRemaining, txHash)
	scrRefund, consumedFee, err := sc.refundGasToSender(totalGasRefund, tx, txHash, acntSnd)
	if err != nil {
		return nil, nil, err
	}

	if scrRefund != nil {
		scrTxs = append(scrTxs, scrRefund)
	}

	err = sc.deleteAccounts(vmOutput.DeletedAccounts)
	if err != nil {
		return nil, nil, err
	}

	err = sc.processTouchedAccounts(vmOutput.TouchedAccounts)
	if err != nil {
		return nil, nil, err
	}

	return scrTxs, consumedFee, nil
}

// reloadLocalSndAccount will reload from current account state the sender account
// this requirement is needed because in the case of refunding the exact account that was previously
// modified in saveSCOutputToCurrentState, the modifications done there should be visible here
func (sc *scProcessor) reloadLocalSndAccount(acntSnd state.AccountHandler) (state.AccountHandler, error) {
	if acntSnd == nil || acntSnd.IsInterfaceNil() {
		return acntSnd, nil
	}

	isAccountFromCurrentShard := acntSnd.AddressContainer() != nil
	if !isAccountFromCurrentShard {
		return acntSnd, nil
	}

	return sc.getAccountFromAddress(acntSnd.AddressContainer().Bytes())
}

func (sc *scProcessor) createSmartContractResult(
	outAcc *vmcommon.OutputAccount,
	scAddress []byte,
	txHash []byte,
) *smartContractResult.SmartContractResult {
	result := &smartContractResult.SmartContractResult{}

	result.Value = outAcc.BalanceDelta
	result.Nonce = outAcc.Nonce
	result.RcvAddr = outAcc.Address
	result.SndAddr = scAddress
	result.Code = outAcc.Code
	result.Data = sc.argsParser.CreateDataFromStorageUpdate(outAcc.StorageUpdates)
	result.TxHash = txHash

	return result
}

func (sc *scProcessor) createSCRTransactions(
	outAccs []*vmcommon.OutputAccount,
	tx *transaction.Transaction,
	txHash []byte,
) ([]data.TransactionHandler, error) {
	scResults := make([]data.TransactionHandler, 0)

	for i := 0; i < len(outAccs); i++ {
		scTx := sc.createSmartContractResult(outAccs[i], tx.RcvAddr, txHash)
		scResults = append(scResults, scTx)
	}

	return scResults, nil
}

// give back the user the unused gas money
func (sc *scProcessor) refundGasToSender(
	gasRefund *big.Int,
	tx *transaction.Transaction,
	txHash []byte,
	acntSnd state.AccountHandler,
) (*smartContractResult.SmartContractResult, *big.Int, error) {
	consumedFee := big.NewInt(0)
	consumedFee = consumedFee.Mul(big.NewInt(0).SetUint64(tx.GasPrice), big.NewInt(0).SetUint64(tx.GasLimit))
	if gasRefund == nil || gasRefund.Cmp(big.NewInt(0)) <= 0 {
		return nil, consumedFee, nil
	}

	refundErd := big.NewInt(0)
	refundErd = refundErd.Mul(gasRefund, big.NewInt(int64(tx.GasPrice)))
	consumedFee = consumedFee.Sub(consumedFee, refundErd)

	scTx := &smartContractResult.SmartContractResult{}
	scTx.Value = refundErd
	scTx.RcvAddr = tx.SndAddr
	scTx.SndAddr = tx.RcvAddr
	scTx.Nonce = tx.Nonce + 1
	scTx.TxHash = txHash

	if acntSnd == nil || acntSnd.IsInterfaceNil() {
		return scTx, consumedFee, nil
	}

	stAcc, ok := acntSnd.(*state.Account)
	if !ok {
		return nil, nil, process.ErrWrongTypeAssertion
	}

	newBalance := big.NewInt(0).Add(stAcc.Balance, refundErd)
	err := stAcc.SetBalanceWithJournal(newBalance)
	if err != nil {
		return nil, nil, err
	}

	return scTx, consumedFee, nil
}

// save account changes in state from vmOutput - protected by VM - every output can be treated as is.
func (sc *scProcessor) processSCOutputAccounts(outputAccounts []*vmcommon.OutputAccount, tx *transaction.Transaction) error {
	sumOfAllDiff := big.NewInt(0)
	sumOfAllDiff = sumOfAllDiff.Sub(sumOfAllDiff, tx.Value)

	zero := big.NewInt(0)
	for i := 0; i < len(outputAccounts); i++ {
		outAcc := outputAccounts[i]
		acc, err := sc.getAccountFromAddress(outAcc.Address)
		if err != nil {
			return err
		}

		if acc == nil || acc.IsInterfaceNil() {
			if outAcc.BalanceDelta != nil {
				sumOfAllDiff = sumOfAllDiff.Add(sumOfAllDiff, outAcc.BalanceDelta)
			}
			continue
		}

		for j := 0; j < len(outAcc.StorageUpdates); j++ {
			storeUpdate := outAcc.StorageUpdates[j]
			acc.DataTrieTracker().SaveKeyValue(storeUpdate.Offset, storeUpdate.Data)
		}

		if len(outAcc.StorageUpdates) > 0 {
			//SC with data variables
			err := sc.accounts.SaveDataTrie(acc)
			if err != nil {
				return err
			}
		}

		// change code if there is a change
		if len(outAcc.Code) > 0 {
			err = sc.accounts.PutCode(acc, outAcc.Code)
			if err != nil {
				return err
			}

			log.Debug("created SC address", "address", hex.EncodeToString(outAcc.Address))
		}

		// change nonce only if there is a change
		if outAcc.Nonce != acc.GetNonce() {
			if outAcc.Nonce < acc.GetNonce() {
				return process.ErrWrongNonceInVMOutput
			}

			err = acc.SetNonceWithJournal(outAcc.Nonce)
			if err != nil {
				return err
			}
		}

		// if no change then continue
		if outAcc.BalanceDelta == nil || outAcc.BalanceDelta.Cmp(zero) == 0 {
			continue
		}

		stAcc, ok := acc.(*state.Account)
		if !ok {
			return process.ErrWrongTypeAssertion
		}

		sumOfAllDiff = sumOfAllDiff.Add(sumOfAllDiff, outAcc.BalanceDelta)

		// update the values according to SC output
		updatedBalance := big.NewInt(0)
		updatedBalance = updatedBalance.Add(stAcc.Balance, outAcc.BalanceDelta)
		if updatedBalance.Cmp(big.NewInt(0)) < 0 {
			return process.ErrOverallBalanceChangeFromSC
		}

		err = stAcc.SetBalanceWithJournal(updatedBalance)
		if err != nil {
			return err
		}
	}

	if sumOfAllDiff.Cmp(zero) != 0 {
		return process.ErrOverallBalanceChangeFromSC
	}

	return nil
}

// delete accounts - only suicide by current SC or another SC called by current SC - protected by VM
func (sc *scProcessor) deleteAccounts(deletedAccounts [][]byte) error {
	for _, value := range deletedAccounts {
		acc, err := sc.getAccountFromAddress(value)
		if err != nil {
			return err
		}

		if acc == nil || acc.IsInterfaceNil() {
			//TODO: sharded Smart Contract processing
			continue
		}

		err = sc.accounts.RemoveAccount(acc.AddressContainer())
		if err != nil {
			return err
		}
	}
	return nil
}

func (sc *scProcessor) processTouchedAccounts(touchedAccounts [][]byte) error {
	//TODO: implement
	return nil
}

func (sc *scProcessor) getAccountFromAddress(address []byte) (state.AccountHandler, error) {
	adrSrc, err := sc.adrConv.CreateAddressFromPublicKeyBytes(address)
	if err != nil {
		return nil, err
	}

	shardForCurrentNode := sc.shardCoordinator.SelfId()
	shardForSrc := sc.shardCoordinator.ComputeId(adrSrc)
	if shardForCurrentNode != shardForSrc {
		return nil, nil
	}

	acnt, err := sc.accounts.GetAccountWithJournal(adrSrc)
	if err != nil {
		return nil, err
	}

	return acnt, nil
}

// GetAllSmartContractCallRootHash returns the roothash of the state of the SC executions for defined round
func (sc *scProcessor) GetAllSmartContractCallRootHash(round uint64) []byte {
	return []byte("roothash")
}

// saves VM output into state
func (sc *scProcessor) saveSCOutputToCurrentState(output *vmcommon.VMOutput, round uint64, txHash []byte) error {
	var err error

	sc.mutSCState.Lock()
	defer sc.mutSCState.Unlock()

	/*
		if _, ok := sc.mapExecState[round]; !ok {
			sc.mapExecState[round] = scExecutionState{
				allLogs:       make(map[string][]*vmcommon.LogEntry),
				allReturnData: make(map[string][]*big.Int),
				returnCodes:   make(map[string]vmcommon.ReturnCode)}
		}*/

	//tmpCurrScState := sc.mapExecState[round]
	defer func() {
		if err != nil {
			//sc.mapExecState[round] = tmpCurrScState
		}
	}()

	err = sc.saveReturnData(output.ReturnData, round, txHash)
	if err != nil {
		return err
	}

	err = sc.saveReturnCode(output.ReturnCode, round, txHash)
	if err != nil {
		return err
	}

	err = sc.saveLogsIntoState(output.Logs, round, txHash)
	if err != nil {
		return err
	}

	return nil
}

// saves return data into account state
func (sc *scProcessor) saveReturnData(returnData [][]byte, round uint64, txHash []byte) error {
	//sc.mapExecState[round].allReturnData[string(txHash)] = returnData
	return nil
}

// saves smart contract return code into account state
func (sc *scProcessor) saveReturnCode(returnCode vmcommon.ReturnCode, round uint64, txHash []byte) error {
	//sc.mapExecState[round].returnCodes[string(txHash)] = returnCode
	return nil
}

// save vm output logs into accounts
func (sc *scProcessor) saveLogsIntoState(logs []*vmcommon.LogEntry, round uint64, txHash []byte) error {
	//sc.mapExecState[round].allLogs[string(txHash)] = logs
	return nil
}

// ProcessSmartContractResult updates the account state from the smart contract result
func (sc *scProcessor) ProcessSmartContractResult(scr *smartContractResult.SmartContractResult) error {
	if scr == nil {
		return process.ErrNilSmartContractResult
	}

	accHandler, err := sc.getAccountFromAddress(scr.RcvAddr)
	if err != nil {
		return err
	}
	if accHandler == nil || accHandler.IsInterfaceNil() {
		return process.ErrNilSCDestAccount
	}

	stAcc, ok := accHandler.(*state.Account)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	storageUpdates, _ := sc.argsParser.GetStorageUpdates(scr.Data)
	for i := 0; i < len(storageUpdates); i++ {
		stAcc.DataTrieTracker().SaveKeyValue(storageUpdates[i].Offset, storageUpdates[i].Data)
	}

	if len(scr.Data) > 0 {
		//SC with data variables
		err := sc.accounts.SaveDataTrie(stAcc)
		if err != nil {
			return err
		}
	}

	if len(scr.Code) > 0 {
		err = sc.accounts.PutCode(stAcc, scr.Code)
		if err != nil {
			return err
		}
	}

	if scr.Value == nil {
		return process.ErrNilBalanceFromSC
	}

	operation := big.NewInt(0)
	operation = operation.Add(scr.Value, stAcc.Balance)
	err = stAcc.SetBalanceWithJournal(operation)
	if err != nil {
		return err
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (sc *scProcessor) IsInterfaceNil() bool {
	if sc == nil {
		return true
	}
	return false
}
