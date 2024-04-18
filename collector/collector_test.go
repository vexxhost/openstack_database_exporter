package collector

import (
	"os"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-kit/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type BaseOpenStackTestSuite struct {
	suite.Suite

	db     *gorm.DB
	logger log.Logger
	mock   sqlmock.Sqlmock
}

func (suite *BaseOpenStackTestSuite) SetupTest() {
	conn, mock, err := sqlmock.New()
	require.NoError(suite.T(), err)

	db, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      conn,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	require.NoError(suite.T(), err)

	suite.db = db
	suite.logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	suite.mock = mock
}

func (suite *BaseOpenStackTestSuite) TearDownTest() {
	assert.NoError(suite.T(), suite.mock.ExpectationsWereMet())
}

func TestOpenStackSuites(t *testing.T) {
	suite.Run(t, &NeutronTestSuite{BaseOpenStackTestSuite: BaseOpenStackTestSuite{}})
	suite.Run(t, &NovaTestSuite{BaseOpenStackTestSuite: BaseOpenStackTestSuite{}})
	suite.Run(t, &OctaviaTestSuite{BaseOpenStackTestSuite: BaseOpenStackTestSuite{}})
}
