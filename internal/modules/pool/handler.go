package pool

import (
	"github.com/gin-gonic/gin"
	"github.com/nodus-protocol/backend/internal/utils"
	"go.uber.org/zap"
)

type Handler struct {
	svc *Service
	log *zap.Logger
}

func NewHandler(svc *Service, log *zap.Logger) *Handler {
	return &Handler{svc: svc, log: log}
}

// GetReserves godoc
// @Summary      Get current AMM pool reserves
// @Tags         pool
// @Produce      json
// @Success      200 {object} utils.Response
// @Router       /pool/reserves [get]
func (h *Handler) GetReserves(c *gin.Context) {
	r, err := h.svc.GetReserves(c.Request.Context())
	if err != nil {
		h.log.Error("get reserves failed", zap.Error(err))
		utils.InternalServerError(c, "failed to fetch pool reserves: "+err.Error())
		return
	}
	utils.OK(c, "reserves retrieved", r)
}

// GetQuote godoc
// @Summary      Get a price quote for a token swap
// @Tags         pool
// @Produce      json
// @Param        amount_in query string true  "Amount of input token (in base units)"
// @Param        token_in  query string true  "Input token symbol (e.g. XLM)"
// @Success      200 {object} utils.Response
// @Router       /pool/quote [get]
func (h *Handler) GetQuote(c *gin.Context) {
	amountIn := c.Query("amount_in")
	tokenIn := c.Query("token_in")
	if amountIn == "" || tokenIn == "" {
		utils.BadRequest(c, "MISSING_PARAMS", "amount_in and token_in are required", nil)
		return
	}

	quote, err := h.svc.GetQuote(c.Request.Context(), amountIn, tokenIn)
	if err != nil {
		utils.InternalServerError(c, "quote failed: "+err.Error())
		return
	}
	utils.OK(c, "quote retrieved", quote)
}

// GetLPBalance godoc
// @Summary      Get LP token balance for a wallet address
// @Tags         pool
// @Produce      json
// @Param        address query string true "Stellar wallet address"
// @Success      200 {object} utils.Response
// @Router       /pool/lp-balance [get]
func (h *Handler) GetLPBalance(c *gin.Context) {
	address := c.Query("address")
	if address == "" {
		utils.BadRequest(c, "MISSING_PARAMS", "address is required", nil)
		return
	}

	bal, err := h.svc.GetLPBalance(c.Request.Context(), address)
	if err != nil {
		utils.InternalServerError(c, "lp balance failed: "+err.Error())
		return
	}
	utils.OK(c, "lp balance retrieved", bal)
}

// GetStats godoc
// @Summary      Get AMM pool statistics (reserves, price, k-invariant)
// @Tags         pool
// @Produce      json
// @Success      200 {object} utils.Response
// @Router       /pool/stats [get]
func (h *Handler) GetStats(c *gin.Context) {
	stats, err := h.svc.GetStats(c.Request.Context())
	if err != nil {
		utils.InternalServerError(c, "pool stats failed: "+err.Error())
		return
	}
	utils.OK(c, "pool stats retrieved", stats)
}

// BuildSwapParams godoc
// @Summary      Build unsigned swap transaction parameters for client-side signing
// @Tags         pool
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body swapParamsRequest true "Swap parameters"
// @Success      200 {object} utils.Response
// @Router       /pool/build/swap [post]
func (h *Handler) BuildSwapParams(c *gin.Context) {
	var req swapParamsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "INVALID_BODY", "malformed request body", nil)
		return
	}

	tx, err := h.svc.BuildSwapParams(c.Request.Context(), req)
	if err != nil {
		utils.InternalServerError(c, "build swap failed: "+err.Error())
		return
	}
	utils.OK(c, "swap transaction parameters ready for signing", tx)
}

// BuildAddLiquidity godoc
// @Summary      Build unsigned add-liquidity transaction parameters
// @Tags         pool
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body addLiquidityRequest true "Add liquidity parameters"
// @Success      200 {object} utils.Response
// @Router       /pool/build/add-liquidity [post]
func (h *Handler) BuildAddLiquidity(c *gin.Context) {
	var req addLiquidityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "INVALID_BODY", "malformed request body", nil)
		return
	}

	tx, err := h.svc.BuildAddLiquidity(c.Request.Context(), req)
	if err != nil {
		utils.InternalServerError(c, "build add-liquidity failed: "+err.Error())
		return
	}
	utils.OK(c, "add-liquidity transaction parameters ready for signing", tx)
}

// BuildRemoveLiquidity godoc
// @Summary      Build unsigned remove-liquidity transaction parameters
// @Tags         pool
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body removeLiquidityRequest true "Remove liquidity parameters"
// @Success      200 {object} utils.Response
// @Router       /pool/build/remove-liquidity [post]
func (h *Handler) BuildRemoveLiquidity(c *gin.Context) {
	var req removeLiquidityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "INVALID_BODY", "malformed request body", nil)
		return
	}

	tx, err := h.svc.BuildRemoveLiquidity(c.Request.Context(), req)
	if err != nil {
		utils.InternalServerError(c, "build remove-liquidity failed: "+err.Error())
		return
	}
	utils.OK(c, "remove-liquidity transaction parameters ready for signing", tx)
}

// GetSnapshots godoc
// @Summary      Get recent cached pool reserve snapshots
// @Tags         pool
// @Produce      json
// @Success      200 {object} utils.Response
// @Router       /pool/snapshots [get]
func (h *Handler) GetSnapshots(c *gin.Context) {
	snaps, err := h.svc.RecentSnapshots(c.Request.Context(), 50)
	if err != nil {
		utils.InternalServerError(c, "snapshots failed: "+err.Error())
		return
	}
	utils.OK(c, "snapshots retrieved", snaps)
}
