package models

// Endpoints is a list of endpoints in a YAML file and all associated data
type Endpoints struct {
	OperationID string          // ie. CosmosBankV1Beta1AllBalances
	Description string          // ie. Query all balances returns all the balances of a single account.
	Tags        []string        // ie /cosmos/bank/v1beta1/balances/{address} = cosmos,bank,balances
	Action      string          // ie. GET, POST, PUT, DELETE
	Path        string          // ie. /cosmos/bank/v1beta1/balances/{address}
	Responses   []string        // ie. 200, 400, 404
	Parameters  []YAMLParameter // ie. address, denom
}
