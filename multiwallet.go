package multiwallet

import (
	"errors"
	"fmt"
	"github.com/cpacia/multiwallet/base"
	"github.com/cpacia/multiwallet/bitcoincash"
	iwallet "github.com/cpacia/wallet-interface"
	"github.com/natefinch/lumberjack"
	"github.com/op/go-logging"
	"os"
	"path"
	"strings"
)

var (
	defaultLogFilename = "multiwallet.log"
	ErrUnsuppertedCoin = errors.New("multiwallet does not contain an implementation for the given coin")
	fileLogFormat      = logging.MustStringFormatter(`%{time:2006-01-02 T15:04:05.000} [%{level}] [%{module}] %{message}`)
	stdoutLogFormat    = logging.MustStringFormatter(`%{color:reset}%{color}%{time:15:04:05} [%{level}] [%{module}] %{message}`)
	logLevelMap        = map[string]logging.Level{
		"debug":    logging.DEBUG,
		"info":     logging.INFO,
		"notice":   logging.NOTICE,
		"warning":  logging.WARNING,
		"error":    logging.ERROR,
		"critical": logging.CRITICAL,
	}
)

type Multiwallet map[iwallet.CoinType]iwallet.Wallet

func NewMultiwallet(cfg *Config) (Multiwallet, error) {
	logger := logging.MustGetLogger("multiwallet")

	backendStdout := logging.NewLogBackend(os.Stdout, "", 0)
	backendStdoutFormatter := logging.NewBackendFormatter(backendStdout, stdoutLogFormat)

	if cfg.LogLevel == "" {
		cfg.LogLevel = "info"
	}
	if cfg.LogDir != "" {
		rotator := &lumberjack.Logger{
			Filename:   path.Join(cfg.LogDir, defaultLogFilename),
			MaxSize:    10, // Megabytes
			MaxBackups: 3,
			MaxAge:     30, // Days
		}

		backendFile := logging.NewLogBackend(rotator, "", 0)
		backendFileFormatter := logging.NewBackendFormatter(backendFile, fileLogFormat)
		leveledBackend := logging.MultiLogger(backendStdoutFormatter, backendFileFormatter)
		leveledBackend.SetLevel(logLevelMap[strings.ToLower(cfg.LogLevel)], "")
		logger.SetBackend(leveledBackend)
	} else {
		leveledBackend := logging.AddModuleLevel(backendStdoutFormatter)
		leveledBackend.SetLevel(logLevelMap[strings.ToLower(cfg.LogLevel)], "")
		logger.SetBackend(leveledBackend)
	}

	multiwallet := make(map[iwallet.CoinType]iwallet.Wallet)
	for _, coinType := range cfg.Wallets {
		switch coinType {
		case iwallet.CtBitcoinCash:
			clientUrl := "bchd.greyh.at:8335"
			if cfg.UseTestnet {
				clientUrl = "testnet-bchd.greyh.at:18335"
			}
			w, err := bitcoincash.NewBitcoinCashWallet(&base.WalletConfig{
				Logger:    logger,
				DataDir:   cfg.DataDir,
				ClientUrl: clientUrl,
				Testnet:   cfg.UseTestnet,
			})
			if err != nil {
				return nil, err
			}

			multiwallet[coinType] = w

		default:
			return nil, fmt.Errorf("a wallet implementation for %s does not exist", coinType.CurrencyCode())
		}
	}

	return multiwallet, nil
}

func (w *Multiwallet) Start() error {
	for _, wallet := range *w {
		if err := wallet.OpenWallet(); err != nil {
			return err
		}
	}
	return nil
}

func (w *Multiwallet) Close() error {
	for _, wallet := range *w {
		if err := wallet.CloseWallet(); err != nil {
			return err
		}
	}
	return nil
}

func (w *Multiwallet) WalletForCurrencyCode(currencyCode string) (iwallet.Wallet, error) {
	for cc, wl := range *w {
		if strings.ToUpper(cc.CurrencyCode()) == strings.ToUpper(currencyCode) || strings.ToUpper(cc.CurrencyCode()) == "T"+strings.ToUpper(currencyCode) {
			return wl, nil
		}
	}
	return nil, ErrUnsuppertedCoin
}