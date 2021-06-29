package conf

import (
	"bytes"
	"fmt"
	"io"
	"net/url"

	"github.com/bketelsen/crypt/backend"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"github.com/pkg/errors"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
)

type NacosBackend struct {
	group        string
	dataID       string
	configClient config_client.IConfigClient
}

type remoteConfigProvider struct {
	originalFactory remoteConfigFactory
}

type remoteConfigFactory interface {
	Get(rp viper.RemoteProvider) (io.Reader, error)
	Watch(rp viper.RemoteProvider) (io.Reader, error)
	WatchChannel(rp viper.RemoteProvider) (<-chan *viper.RemoteResponse, chan bool)
}

func MakeVipperSupportNacos() {
	if _, ok := viper.RemoteConfig.(*remoteConfigProvider); ok {
		return
	}
	viper.SupportedRemoteProviders = append(viper.SupportedRemoteProviders, "nacos")
	viper.RemoteConfig = &remoteConfigProvider{
		originalFactory: viper.RemoteConfig,
	}
}

func NewNacosBackend(machines []string) (*NacosBackend, error) {
	if len(machines) != 1 {
		return nil, fmt.Errorf("nacos target invalid")
	}
	//remain := machines[0]
	//index := strings.Index(remain, ":")
	//if index <= 0 {
	//	return nil, fmt.Errorf("nacos uri invalid,not contains user")
	//}
	//user := remain[:index]
	//remain = remain[index+1:]
	//index = strings.Index(remain, "@")
	//if index <= 0 {
	//	return nil, fmt.Errorf("nacos uri invalid,not contains pass")
	//}
	//password := remain[:index]
	//remain = remain[index+1:]
	//index = strings.Index(remain, "/")
	//if index <= 0 {
	//	return nil, fmt.Errorf("nacos uri invalid,not contains namespace")
	//}
	//namespace := remain[:index]
	u, err := url.Parse(machines[0])
	if err != nil {
		return nil, errors.WithMessage(err, "nacos uri invalid")
	}
	if u.Scheme != "nacos" {
		return nil, fmt.Errorf("nacos uri schema invalid")
	}
	if u.User == nil {
		return nil, fmt.Errorf("nacos uri userinfo invalid")
	}

	if u.User.Username() == "" {
		return nil, fmt.Errorf("nacos uri user invalid")
	}
	password, ok := u.User.Password()
	if !ok {
		return nil, fmt.Errorf("nacos uri password invalid")
	}
	query := u.Query()
	if len(query["namespace"]) != 1 {
		return nil, fmt.Errorf("nacos namespace invalid")
	}
	if len(query["dataID"]) != 1 {
		return nil, fmt.Errorf("nacos dataID invalid")
	}

	group := "DEFAULT_GROUP"
	if len(query["group"]) == 1 {
		group = query["group"][0]
	}

	var timeout uint64 = 5000
	if len(query["timeout"]) == 1 {
		timeout = cast.ToUint64(u.Query()["timeout"][0])
	}

	clientConfig := constant.ClientConfig{
		Username:            u.User.Username(),
		Password:            password,
		NamespaceId:         u.Query()["namespace"][0],
		TimeoutMs:           timeout,
		NotLoadCacheAtStart: true,
	}
	var port uint64 = 80
	if u.Port() != "" {
		port = cast.ToUint64(u.Port())
	}
	serverConfigs := []constant.ServerConfig{
		{
			IpAddr:      u.Hostname(),
			ContextPath: u.Path,
			Port:        port,
		},
	}

	configClient, err := clients.CreateConfigClient(map[string]interface{}{
		"serverConfigs": serverConfigs,
		"clientConfig":  clientConfig,
	})

	if err != nil {
		return nil, errors.WithMessage(err, "create nacos client")
	}

	return &NacosBackend{
		group:        group,
		dataID:       query["dataID"][0],
		configClient: configClient,
	}, nil
}

func (self *NacosBackend) Get(key string) ([]byte, error) {
	confData, err := self.configClient.GetConfig(vo.ConfigParam{
		DataId: self.dataID,
		Group:  self.group,
	})

	return []byte(confData), err
}

func (self *NacosBackend) List(key string) (backend.KVPairs, error) {
	panic("implement me")
}

func (self *NacosBackend) Set(key string, value []byte) error {
	_, err := self.configClient.PublishConfig(vo.ConfigParam{
		DataId:  self.dataID,
		Group:   self.group,
		Content: string(value),
	})
	return err
}

func (self *NacosBackend) Watch(key string, stop chan bool) <-chan *backend.Response {
	panic("implement me")
}

func (self *remoteConfigProvider) Get(rp viper.RemoteProvider) (io.Reader, error) {
	if rp.Provider() != "nacos" {
		return self.originalFactory.Get(rp)
	} else {
		cm, err := NewNacosBackend([]string{rp.Endpoint()})
		if err != nil {
			return nil, err
		}
		data, err := cm.Get(rp.Path())
		if err != nil {
			return nil, err
		}
		return bytes.NewBuffer(data), nil
	}
}

func (self *remoteConfigProvider) Watch(rp viper.RemoteProvider) (io.Reader, error) {
	if rp.Provider() != "nacos" {
		return self.originalFactory.Get(rp)
	} else {
		cm, err := NewNacosBackend([]string{rp.Endpoint()})
		if err != nil {
			return nil, err
		}
		data, err := cm.Get(rp.Path())
		if err != nil {
			return nil, err
		}
		return bytes.NewBuffer(data), nil
	}
}

func (self *remoteConfigProvider) WatchChannel(rp viper.RemoteProvider) (<-chan *viper.RemoteResponse, chan bool) {
	panic("implement me")
}
