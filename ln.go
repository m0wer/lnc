package lnc

import "errors"

var (
	PaymentHashExists = errors.New("invoice with that payment hash already exists")
	PaymentFailed     = errors.New("payment failed")
)

type LN interface {
	DecodeInvoice(string) (*DecodedInvoice, error)
	// If InvoiceParameters.Hash is not nil, invoice will be a hold invoice,
	// which must be settled with SettleInvoice
	AddInvoice(InvoiceParameters) (string, error)
	WatchInvoice([]byte) (*InvoiceState, error)
	CancelInvoice([]byte) error
	// If `error == nil`, the payment succeeded,
	// else if `errors.Is(error, lnc.PaymentFailed)`, the payment failed,
	// otherwise the payment status is unknown
	PayInvoice(PaymentParameters) ([]byte, error)
	SettleInvoice([]byte) error
	// Lower bound routing fee and cltv_delta estimate to pay the invoice
	EstimateRoutingFee(DecodedInvoice, uint64) (min_fee_msat uint64, min_cltv_delta uint64, err error)
}

// Signer is implemented by backends that can prove control of the node's
// identity key. It is used by lnproxy to attest that a nostr advertisement
// belongs to a particular lightning node. It is kept separate from LN so that
// backends without message-signing support can still implement LN.
type Signer interface {
	// IdentityPubkey returns the hex-encoded compressed identity public key of
	// the node.
	IdentityPubkey() (string, error)
	// SignMessage signs msg with the node's identity key and returns a
	// zbase32-encoded recoverable signature, as produced by LND's
	// `lnrpc.Lightning/SignMessage` and `lncli signmessage`.
	SignMessage(msg []byte) (string, error)
}

type DecodedInvoice struct {
	PaymentHash     string `json:"payment_hash"`
	Timestamp       uint64 `json:"timestamp,string"`
	Expiry          uint64 `json:"expiry,string"`
	Description     string `json:"description"`
	DescriptionHash string `json:"description_hash"`
	NumMsat         uint64 `json:"num_msat,string"`
	CltvExpiry      uint64 `json:"cltv_expiry,string"`
	Features        map[string]struct {
		Name       string `json:"name"`
		IsRequired bool   `json:"is_required"`
		IsKnown    bool   `json:"is_known"`
	} `json:"features"`
	Destination string      `json:"destination"`
	RouteHints  []RouteHint `json:"route_hints"`
}

type RouteHint struct {
	HopHints []HopHint `json:"hop_hints"`
}

type HopHint struct {
	NodeId          string `json:"node_id"`
	ChanId          uint64 `json:"chan_id,string"`
	FeeBaseMsat     uint64 `json:"fee_base_msat"`
	FeePPM          uint64 `json:"fee_proportional_millionths"`
	CltvExpiryDelta uint64 `json:"cltv_expiry_delta"`
}

type InvoiceParameters struct {
	Memo            string `json:"memo,omitempty"`
	Hash            []byte `json:"hash,omitempty"`
	ValueMsat       uint64 `json:"value_msat,string"`
	DescriptionHash []byte `json:"description_hash,omitempty"`
	Expiry          uint64 `json:"expiry,string"`
	CltvExpiry      uint64 `json:"cltv_expiry,string"`
}

type InvoiceState struct {
	State
	AmtPaid         uint64
	CltvExpiryDelta uint64
}

type State int

const (
	Unknown State = iota
	Canceled
	Accepted
	Settled
)

type PaymentParameters struct {
	Invoice        string `json:"payment_request"`
	AmtMsat        uint64 `json:"amt_msat,omitempty,string"`
	TimeoutSeconds uint64 `json:"timeout_seconds"`
	FeeLimitMsat   uint64 `json:"fee_limit_msat,string"`
	CltvLimit      uint64 `json:"cltv_limit"`
}
