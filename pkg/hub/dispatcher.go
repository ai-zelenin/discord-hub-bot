package hub

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/antonmedv/expr"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"text/template"
)

type Dispatcher struct {
	state  *FileStore
	bot    *Bot
	cfg    *Config
	logger *log.Logger
}

func NewDispatcher(logger *log.Logger, cfg *Config, state *FileStore, bot *Bot) *Dispatcher {
	return &Dispatcher{
		state:  state,
		bot:    bot,
		cfg:    cfg,
		logger: logger,
	}
}

func (d *Dispatcher) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := d.handleRequest(w, r)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
	_, _ = fmt.Fprint(w, "ok")
}

func (d *Dispatcher) handleRequest(_ http.ResponseWriter, r *http.Request) error {
	key := strings.TrimPrefix(r.RequestURI, "/")
	d.logger.Printf("income on key %s", key)
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	chCfg, ok := d.cfg.Sources()[key]
	if ok {
		var params interface{}
		switch chCfg.RequestType {
		case RequestTypePostJsonObj:
			params = make(map[string]interface{})
			err = json.Unmarshal(body, &params)
			if err != nil {
				return err
			}
		case RequestTypePostJsonArray:
			params = make([]map[string]interface{}, 0)
			err = json.Unmarshal(body, &params)
			if err != nil {
				return err
			}
		case RequestTypeGet:
			getParams := make(map[string]string)
			for k, vals := range r.URL.Query() {
				getParams[k] = strings.Join(vals, ",")
			}
			params = getParams
		default:
			return fmt.Errorf("unkown RequestType [%s]", chCfg.RequestType)
		}
		err = d.HandlerWebhook(key, body, params, chCfg)
		if err != nil {
			return err
		}
	} else {
		d.logger.Println(string(body))
	}
	return nil

}

func (d *Dispatcher) HandlerWebhook(source string, body []byte, params interface{}, chCfg *SourceConfig) error {
	d.logger.Printf("handle webhook source:%s", source)
	return d.Proxy(source, body, params, chCfg)
}

func (d *Dispatcher) Proxy(source string, body []byte, params interface{}, srcCfg *SourceConfig) error {
	tp := template.New(source)
	tp = tp.Funcs(template.FuncMap{
		"mention": func(userID string) string {
			return fmt.Sprintf("<@%s>", userID)
		},
	})
	// TODO optimize
	tp, err := tp.Parse(srcCfg.Template)
	if err != nil {
		return err
	}

	subs, err := d.state.FindSubscriptionsBySource(source)
	if err != nil {
		return err
	}
	for _, sub := range subs {
		scope := &Scope{
			Raw: string(body),
			R:   params,
			Sub: sub,
			Cfg: srcCfg,
		}

		if sub.OnlyPersonal {
			ok, err := d.isPersonal(scope)
			if err != nil {
				d.sendMessage(fmt.Sprintf("expression exec failure err:%v", err), sub)
				continue
			}
			if !ok {
				continue
			}
		}

		ok, err := d.filter(sub, scope)
		if err != nil {
			d.sendMessage(fmt.Sprintf("expression exec failure err:%v", err), sub)
			continue
		}
		if ok {
			msgBody := bytes.NewBuffer(nil)
			err = tp.Execute(msgBody, scope)
			if err != nil {
				return err
			}
			d.sendMessage(msgBody.String(), sub)
		}
	}
	return nil
}

func (d *Dispatcher) sendMessage(msg string, sub *Subscription) {
	if sub.Direct {
		d.bot.SendMessageToDirect(msg, sub.UserID)
	} else {
		d.bot.SendMessage(msg, sub.ChannelID)
	}
}

func (d *Dispatcher) isPersonal(scope *Scope) (bool, error) {
	if scope.Cfg.FindPersonaExpr == "" {
		return true, nil
	}
	// todo optimize
	program, err := expr.Compile(scope.Cfg.FindPersonaExpr, expr.Env(scope))
	if err != nil {
		return false, err
	}
	output, err := expr.Run(program, scope)
	if err != nil {
		return false, err
	}
	result, ok := output.(bool)
	if !ok {
		return false, fmt.Errorf("output [%T] is unacceptable bool type is required", output)
	}
	return result, nil
}

func (d *Dispatcher) filter(sub *Subscription, scope *Scope) (bool, error) {
	if sub.CustomFilter == "" {
		return true, nil
	}
	// todo optimize
	program, err := expr.Compile(sub.CustomFilter, expr.Env(scope))
	if err != nil {
		return false, err
	}
	output, err := expr.Run(program, scope)
	if err != nil {
		return false, err
	}
	result, ok := output.(bool)
	if !ok {
		return false, fmt.Errorf("output [%T] is unacceptable bool type is required", output)
	}
	return result, nil
}
