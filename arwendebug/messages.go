package arwendebug

import (
	"encoding/hex"
	"io/ioutil"
	"math/big"

	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

// RequestBase is a CLI / REST request message
type RequestBase struct {
	DatabasePath string
	World        string
	Outcome      string
}

// ResponseBase is a CLI / REST response message
type ResponseBase struct {
	Error error
}

// ContractRequestBase is a CLI / REST request message
type ContractRequestBase struct {
	RequestBase
	Impersonated string
	Value        string
	GasPrice     uint64
	GasLimit     uint64
}

func (request *ContractRequestBase) getValue() *big.Int {
	value := big.NewInt(0)
	_, _ = value.SetString(request.Value, 10)
	return value
}

func (request *ContractRequestBase) getGasPrice() uint64 {
	if request.GasPrice == 0 {
		return DefaultGasPrice
	}

	return request.GasPrice
}

func (request *ContractRequestBase) getGasLimit() uint64 {
	if request.GasLimit == 0 {
		return DefaultGasLimit
	}

	return request.GasLimit
}

func (request *ContractRequestBase) getImpersonated() ([]byte, error) {
	impersonatedAsHex := request.Impersonated
	impersonatedAsBytes, err := hex.DecodeString(impersonatedAsHex)
	if err != nil {
		return nil, NewRequestErrorMessageInner("invalid impersonated address", err)
	}

	return impersonatedAsBytes, nil
}

// ContractResponseBase is a CLI / REST response message
type ContractResponseBase struct {
	ResponseBase
	Input  *vmcommon.VMInput
	Output *vmcommon.VMOutput
}

func (response *ContractResponseBase) getReturnCode() vmcommon.ReturnCode {
	return response.Output.ReturnCode
}

// DeployRequest is a CLI / REST request message
type DeployRequest struct {
	ContractRequestBase
	Code         string
	CodePath     string
	CodeMetadata string
	Arguments    []string
}

func (request *DeployRequest) getCode() ([]byte, error) {
	if len(request.Code) > 0 {
		codeAsHex := request.Code
		codeAsBytes, err := hex.DecodeString(codeAsHex)
		if err != nil {
			return nil, NewRequestErrorMessageInner("invalid contract code", err)
		}

		return codeAsBytes, nil
	}

	if len(request.CodePath) > 0 {
		codeAsBytes, err := ioutil.ReadFile(request.CodePath)
		if err != nil {
			return nil, err
		}

		return codeAsBytes, nil
	}

	return nil, NewRequestError("invalid contract code")
}

func (request *DeployRequest) getCodeMetadata() ([]byte, error) {
	if len(request.CodeMetadata) > 0 {
		metadataAsHex := request.CodeMetadata
		metadataAsBytes, err := hex.DecodeString(metadataAsHex)
		if err != nil {
			return nil, err
		}

		return metadataAsBytes, nil
	}

	defaultMetadata := vmcommon.CodeMetadata{Upgradeable: true}
	return defaultMetadata.ToBytes(), nil
}

func (request *DeployRequest) getArguments() ([][]byte, error) {
	return decodeArguments(request.Arguments)
}

// DeployResponse is a CLI / REST response message
type DeployResponse struct {
	ContractResponseBase
	ContractAddress string
}

// UpgradeRequest is a CLI / REST request message
type UpgradeRequest struct {
	DeployRequest
	ContractAddress string
}

// UpgradeResponse is a CLI / REST response message
type UpgradeResponse struct {
	ContractResponseBase
}

// RunRequest is a CLI / REST request message
type RunRequest struct {
	ContractRequestBase
	ContractAddress string
	Function        string
	Arguments       []string
}

func (request *RunRequest) getArguments() ([][]byte, error) {
	return decodeArguments(request.Arguments)
}

// RunResponse is a CLI / REST response message
type RunResponse struct {
	ContractResponseBase
}

// QueryRequest is a CLI / REST request message
type QueryRequest struct {
	RunRequest
}

// QueryResponse is a CLI / REST response message
type QueryResponse struct {
	ContractResponseBase
}

// CreateAccountRequest is a CLI / REST request message
type CreateAccountRequest struct {
	RequestBase
	Address string
	Balance string
	Nonce   uint64
}

func (request *CreateAccountRequest) getAddress() ([]byte, error) {
	addressAsHex := request.Address
	addressAsBytes, err := hex.DecodeString(addressAsHex)
	if err != nil {
		return nil, NewRequestErrorMessageInner("invalid account address", err)
	}

	return addressAsBytes, nil
}

func (request *CreateAccountRequest) getBalance() (*big.Int, error) {
	balance, ok := big.NewInt(0).SetString(request.Balance, 10)
	if !ok {
		return nil, NewRequestError("invalid balance")
	}

	return balance, nil
}

// CreateAccountResponse is a CLI / REST response message
type CreateAccountResponse struct {
	Account *Account
}
