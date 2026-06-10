package payments

// ── Core Engine request types ─────────────────────────────────────────────────

type initiateRequest struct {
	Sender         string `json:"sender"          binding:"required"`
	Recipient      string `json:"recipient"       binding:"required"`
	Amount         uint64 `json:"amount"          binding:"required,min=1"`
	Token          string `json:"token"           binding:"required"`
	Urgency        string `json:"urgency"`
	IdempotencyKey string `json:"idempotency_key"`
}

type simulateRequest struct {
	Sender    string `json:"sender"    binding:"required"`
	Recipient string `json:"recipient" binding:"required"`
	Amount    uint64 `json:"amount"    binding:"required,min=1"`
	Token     string `json:"token"     binding:"required"`
	Urgency   string `json:"urgency"`
}

type batchItem struct {
	Sender    string `json:"sender"    binding:"required"`
	Recipient string `json:"recipient" binding:"required"`
	Amount    uint64 `json:"amount"    binding:"required,min=1"`
	Token     string `json:"token"     binding:"required"`
	Urgency   string `json:"urgency"`
}

// ── Core Engine response types ────────────────────────────────────────────────

type enginePayment struct {
	ID         string  `json:"id"`
	Sender     string  `json:"sender"`
	Recipient  string  `json:"recipient"`
	Amount     uint64  `json:"amount"`
	Token      string  `json:"token"`
	Status     string  `json:"status"`
	TxHash     *string `json:"tx_hash"`
	FeeStroops uint64  `json:"fee_stroops"`
	Urgency    string  `json:"urgency"`
	Error      *string `json:"error"`
	CreatedAt  string  `json:"created_at"`
	UpdatedAt  string  `json:"updated_at"`
}

type engineSimulation struct {
	Sender                      string `json:"sender"`
	Recipient                   string `json:"recipient"`
	Amount                      uint64 `json:"amount"`
	Token                       string `json:"token"`
	FeeStroops                  uint64 `json:"fee_stroops"`
	Chain                       string `json:"chain"`
	EstimatedConfirmationSeconds uint32 `json:"estimated_confirmation_seconds"`
}

type engineReceipt struct {
	PaymentID   string `json:"payment_id"`
	TxHash      string `json:"tx_hash"`
	Sender      string `json:"sender"`
	Recipient   string `json:"recipient"`
	Amount      uint64 `json:"amount"`
	Token       string `json:"token"`
	Chain       string `json:"chain"`
	ConfirmedAt string `json:"confirmed_at"`
}

type engineBatchItemResult struct {
	Index   int            `json:"index"`
	Payment *enginePayment `json:"payment"`
	Error   *string        `json:"error"`
}

type engineBatchResult struct {
	Total     int                     `json:"total"`
	Succeeded int                     `json:"succeeded"`
	Failed    int                     `json:"failed"`
	Results   []engineBatchItemResult `json:"results"`
}

type engineFees struct {
	Chain     string `json:"chain"`
	Available bool   `json:"available"`
	Fees      struct {
		StandardStroops uint64 `json:"standard_stroops"`
		FastStroops     uint64 `json:"fast_stroops"`
		UrgentStroops   uint64 `json:"urgent_stroops"`
		StandardSeconds uint32 `json:"standard_seconds"`
		FastSeconds     uint32 `json:"fast_seconds"`
		UrgentSeconds   uint32 `json:"urgent_seconds"`
	} `json:"fees"`
}

type engineRate struct {
	Token     string  `json:"token"`
	USDPrice  float64 `json:"usd_price"`
	Available bool    `json:"available"`
}

type engineHealth struct {
	Status          string   `json:"status"`
	Chains          []string `json:"chains"`
	PaymentsInStore int      `json:"payments_in_store"`
}
