package payments

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/OpenNSW/nsw-task-flow/plugins"
)

// TODO: Add payment gateway integration here
type PaymentPlugin struct{}

func NewPaymentPlugin() *PaymentPlugin {
	return &PaymentPlugin{}
}

func (p *PaymentPlugin) Name() string {
	return "PAYMENT_SRILANKA"
}

// Config is the configuration data from the task template's plugin_properties
// TODO: Extend this structure to include actual merchant credentials, gateway URLs, and redirect configurations
type Config struct {
	ItemName string  `json:"item_name"`
	Amount   float64 `json:"amount"`
	VAT      float64 `json:"vat"`
	Total    float64 `json:"total"`
}

func (p *PaymentPlugin) Execute(ctx plugins.PluginContext, configRaw json.RawMessage) error {
	var cfg Config
	if err := json.Unmarshal(configRaw, &cfg); err != nil {
		return fmt.Errorf("failed to parse PAYMENT_SRILANKA config: %w", err)
	}

	ctx.Record.Status = "PENDING_PAYMENT"

	if ctx.Record.Data == nil {
		ctx.Record.Data = make(map[string]any)
	}
	ctx.Record.Data["payment_details"] = cfg

	log.Printf("[Plugin: PAYMENT_SRILANKA] Task %s pending payment. Waiting for external payment confirmation", ctx.Record.TaskID)

	// TODO: Replace this placeholder suspend with actual API invocation to the payment gateway
	// and wait for the successful payment webhook callback to resume the workflow.
	return plugins.ErrSuspended
}
