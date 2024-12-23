package model

type ServiceApi struct {
	ID         string `json:"ID"`
	Node       string `json:"Node"`
	Address    string `json:"Address"`
	Datacenter string `json:"Datacenter"`
	//	TaggedAddresses struct {
	//		Lan     string `json:"lan"`
	//		LanIpv4 string `json:"lan_ipv4"`
	//		Wan     string `json:"wan"`
	//		WanIpv4 string `json:"wan_ipv4"`
	//	} `json:"TaggedAddresses"`
	NodeMeta struct {
		ConsulNetworkSegment string `json:"consul-network-segment"`
		ConsulVersion        string `json:"consul-version"`
	} `json:"NodeMeta"`
	ServiceKind    string `json:"ServiceKind"`
	ServiceID      string `json:"ServiceID"`
	ServiceName    string `json:"ServiceName"`
	ServiceTags    []any  `json:"ServiceTags"`
	ServiceAddress string `json:"ServiceAddress"`
	ServiceWeights struct {
		Passing int `json:"Passing"`
		Warning int `json:"Warning"`
	} `json:"ServiceWeights"`
	ServiceMeta struct {
	} `json:"ServiceMeta"`
	ServicePort              int    `json:"ServicePort"`
	ServiceSocketPath        string `json:"ServiceSocketPath"`
	ServiceEnableTagOverride bool   `json:"ServiceEnableTagOverride"`
	ServiceProxy             struct {
		Mode        string `json:"Mode"`
		MeshGateway struct {
		} `json:"MeshGateway"`
		Expose struct {
		} `json:"Expose"`
	} `json:"ServiceProxy"`
	ServiceConnect struct {
	} `json:"ServiceConnect"`
	ServiceLocality any `json:"ServiceLocality"`
	CreateIndex     int `json:"CreateIndex"`
	ModifyIndex     int `json:"ModifyIndex"`
}

type HealthCheck []struct {
	Node struct {
		ID              string `json:"ID"`
		Node            string `json:"Node"`
		Address         string `json:"Address"`
		Datacenter      string `json:"Datacenter"`
		TaggedAddresses struct {
			Lan     string `json:"lan"`
			LanIpv4 string `json:"lan_ipv4"`
			Wan     string `json:"wan"`
			WanIpv4 string `json:"wan_ipv4"`
		} `json:"TaggedAddresses"`
		Meta struct {
			ConsulNetworkSegment string `json:"consul-network-segment"`
			ConsulVersion        string `json:"consul-version"`
		} `json:"Meta"`
		CreateIndex int `json:"CreateIndex"`
		ModifyIndex int `json:"ModifyIndex"`
	} `json:"Node"`
	Service struct {
		ID      string `json:"ID"`
		Service string `json:"Service"`
		Tags    []any  `json:"Tags"`
		Address string `json:"Address"`
		Meta    any    `json:"Meta"`
		Port    int    `json:"Port"`
		Weights struct {
			Passing int `json:"Passing"`
			Warning int `json:"Warning"`
		} `json:"Weights"`
		EnableTagOverride bool `json:"EnableTagOverride"`
		Proxy             struct {
			Mode        string `json:"Mode"`
			MeshGateway struct {
			} `json:"MeshGateway"`
			Expose struct {
			} `json:"Expose"`
		} `json:"Proxy"`
		Connect struct {
		} `json:"Connect"`
		PeerName    string `json:"PeerName"`
		CreateIndex int    `json:"CreateIndex"`
		ModifyIndex int    `json:"ModifyIndex"`
	} `json:"Service"`
	Checks []struct {
		Node        string `json:"Node"`
		CheckID     string `json:"CheckID"`
		Name        string `json:"Name"`
		Status      string `json:"Status"`
		Notes       string `json:"Notes"`
		Output      string `json:"Output"`
		ServiceID   string `json:"ServiceID"`
		ServiceName string `json:"ServiceName"`
		ServiceTags []any  `json:"ServiceTags"`
		Type        string `json:"Type"`
		Interval    string `json:"Interval"`
		Timeout     string `json:"Timeout"`
		ExposedPort int    `json:"ExposedPort"`
		Definition  struct {
		} `json:"Definition"`
		CreateIndex int `json:"CreateIndex"`
		ModifyIndex int `json:"ModifyIndex"`
	} `json:"Checks"`
}
