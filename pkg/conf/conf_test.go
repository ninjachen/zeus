package conf_test

import (
	"fmt"
	"testing"

	"github.com/spf13/viper"

	"go.yym.plus/zeus/pkg/conf"
)

func TestNacos(t *testing.T) {
	conf.MakeVipperSupportNacos()
	viper.SetConfigType("yaml")
	viper.AddRemoteProvider("nacos", "nacos://nacos:21KkF8c5SCh0@nacos.yym.plus/nacos?namespace=4b8a38fb-de6a-4287-96dd-410d1f923acf&dataID=phz", "")
	err := viper.ReadRemoteConfig()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Print(viper.Get("test"))
}

func TestSource(t *testing.T) {

}