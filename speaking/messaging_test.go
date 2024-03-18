package speaking

import (
	"testing"

	"gopkg.in/tucnak/telebot.v2"
	"ingresos_gastos/storage_interface"
)

func TestEnv_DetectAppropriateActionForInput(t *testing.T) {
	type fields struct {
		Storage storage_interface.ActualStorage
		Bot     *telebot.Bot
	}
	type args struct {
		user        *telebot.User
		messageText string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{ }
	},
		for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				Storage: tt.fields.Storage,
				Bot:     tt.fields.Bot,
			}
			env.DetectAppropriateActionForInput(tt.args.user, tt.args.messageText)
		})
	}
}
