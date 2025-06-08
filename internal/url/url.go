package url

import (
	"github.com/pkg/errors"
	"github.com/samber/do/v2"
	"github.com/sirupsen/logrus"
	"github.com/wickedv43/go-shortener/internal/config"
	"github.com/wickedv43/go-shortener/internal/logger"
	"github.com/wickedv43/go-shortener/internal/storage"
)

type URLService struct {
	cfg     *config.Config     // Configuration settings.
	logger  *logrus.Entry      // Structured logger.
	storage storage.DataKeeper // Interface to storage layer.
}

func NewService(i do.Injector) (*URLService, error) {
	u, err := do.InvokeStruct[URLService](i)
	if err != nil {
		return nil, errors.Wrap(err, "invoke struct error")
	}

	u.cfg = do.MustInvoke[*config.Config](i)
	u.logger = do.MustInvoke[*logger.Logger](i).WithField("component", "url service")
	u.storage = do.MustInvoke[storage.DataKeeper](i)

	return u, nil
}
