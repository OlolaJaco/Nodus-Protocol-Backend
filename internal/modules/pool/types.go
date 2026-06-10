package pool

// ── Core Engine pool response types ──────────────────────────────────────────

type reserves struct {
	Reserve0      string `json:"reserve_0"`
	Reserve1      string `json:"reserve_1"`
	Token0        string `json:"token_0"`
	Token1        string `json:"token_1"`
	LpTotalSupply string `json:"lp_total_supply"`
	TimestampLast uint64 `json:"timestamp_last"`
}

type priceQuote struct {
	AmountIn       string  `json:"amount_in"`
	AmountOut      string  `json:"amount_out"`
	TokenIn        string  `json:"token_in"`
	TokenOut       string  `json:"token_out"`
	FeeBps         uint64  `json:"fee_bps"`
	PriceImpactBps uint64  `json:"price_impact_bps"`
	EffectivePrice float64 `json:"effective_price"`
}

type lpBalance struct {
	Address   string `json:"address"`
	LpBalance string `json:"lp_balance"`
}

type unsignedTx struct {
	ContractID string      `json:"contract_id"`
	Function   string      `json:"function"`
	Args       interface{} `json:"args"`
	Note       string      `json:"note"`
}

type poolStats struct {
	Reserves           reserves `json:"reserves"`
	PriceToken0InToken1 float64 `json:"price_token0_in_token1"`
	PriceToken1InToken0 float64 `json:"price_token1_in_token0"`
	KInvariant         string  `json:"k_invariant"`
	FeeBps             uint64  `json:"fee_bps"`
}

// ── Handler request types ─────────────────────────────────────────────────────

type swapParamsRequest struct {
	To          string `json:"to"           binding:"required"`
	Amount0Out  string `json:"amount_0_out" binding:"required"`
	Amount1Out  string `json:"amount_1_out" binding:"required"`
	Deadline    uint64 `json:"deadline"     binding:"required"`
}

type addLiquidityRequest struct {
	From             string `json:"from"              binding:"required"`
	To               string `json:"to"                binding:"required"`
	Amount0Desired   string `json:"amount_0_desired"  binding:"required"`
	Amount1Desired   string `json:"amount_1_desired"  binding:"required"`
	Amount0Min       string `json:"amount_0_min"      binding:"required"`
	Amount1Min       string `json:"amount_1_min"      binding:"required"`
	Deadline         uint64 `json:"deadline"          binding:"required"`
}

type removeLiquidityRequest struct {
	From       string `json:"from"        binding:"required"`
	To         string `json:"to"          binding:"required"`
	Liquidity  string `json:"liquidity"   binding:"required"`
	Amount0Min string `json:"amount_0_min" binding:"required"`
	Amount1Min string `json:"amount_1_min" binding:"required"`
	Deadline   uint64 `json:"deadline"    binding:"required"`
}
