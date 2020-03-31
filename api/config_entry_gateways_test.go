package api

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAPI_ConfigEntries_IngressGateway(t *testing.T) {
	t.Parallel()
	c, s := makeClient(t)
	defer s.Stop()

	config_entries := c.ConfigEntries()

	ingress1 := &IngressGatewayConfigEntry{
		Kind: IngressGateway,
		Name: "foo",
	}

	ingress2 := &IngressGatewayConfigEntry{
		Kind: IngressGateway,
		Name: "bar",
	}

	// set it
	_, wm, err := config_entries.Set(ingress1, nil)
	require.NoError(t, err)
	require.NotNil(t, wm)
	require.NotEqual(t, 0, wm.RequestTime)

	// also set the second one
	_, wm, err = config_entries.Set(ingress2, nil)
	require.NoError(t, err)
	require.NotNil(t, wm)
	require.NotEqual(t, 0, wm.RequestTime)

	// get it
	entry, qm, err := config_entries.Get(IngressGateway, "foo", nil)
	require.NoError(t, err)
	require.NotNil(t, qm)
	require.NotEqual(t, 0, qm.RequestTime)

	// verify it
	readIngress, ok := entry.(*IngressGatewayConfigEntry)
	require.True(t, ok)
	require.Equal(t, ingress1.Kind, readIngress.Kind)
	require.Equal(t, ingress1.Name, readIngress.Name)

	// update it
	ingress1.Listeners = []IngressListener{
		{
			Port:     2222,
			Protocol: "tcp",
			Services: []IngressService{
				{
					Name: "asdf",
				},
			},
		},
	}

	// CAS fail
	written, _, err := config_entries.CAS(ingress1, 0, nil)
	require.NoError(t, err)
	require.False(t, written)

	// CAS success
	written, wm, err = config_entries.CAS(ingress1, readIngress.ModifyIndex, nil)
	require.NoError(t, err)
	require.NotNil(t, wm)
	require.NotEqual(t, 0, wm.RequestTime)
	require.True(t, written)

	// update no cas
	ingress2.Listeners = []IngressListener{
		{
			Port:     3333,
			Protocol: "http",
			Services: []IngressService{
				{
					Name: "qwer",
				},
			},
		},
	}
	_, wm, err = config_entries.Set(ingress2, nil)
	require.NoError(t, err)
	require.NotNil(t, wm)
	require.NotEqual(t, 0, wm.RequestTime)

	// list them
	entries, qm, err := config_entries.List(IngressGateway, nil)
	require.NoError(t, err)
	require.NotNil(t, qm)
	require.NotEqual(t, 0, qm.RequestTime)
	require.Len(t, entries, 2)

	for _, entry = range entries {
		switch entry.GetName() {
		case "foo":
			// this also verifies that the update value was persisted and
			// the updated values are seen
			readIngress, ok = entry.(*IngressGatewayConfigEntry)
			require.True(t, ok)
			require.Equal(t, ingress1.Kind, readIngress.Kind)
			require.Equal(t, ingress1.Name, readIngress.Name)
			require.Equal(t, ingress1.Listeners, readIngress.Listeners)
		case "bar":
			readIngress, ok = entry.(*IngressGatewayConfigEntry)
			require.True(t, ok)
			require.Equal(t, ingress2.Kind, readIngress.Kind)
			require.Equal(t, ingress2.Name, readIngress.Name)
			require.Equal(t, ingress2.Listeners, readIngress.Listeners)
		}
	}

	// delete it
	wm, err = config_entries.Delete(IngressGateway, "foo", nil)
	require.NoError(t, err)
	require.NotNil(t, wm)
	require.NotEqual(t, 0, wm.RequestTime)

	// verify deletion
	entry, qm, err = config_entries.Get(IngressGateway, "foo", nil)
	require.Error(t, err)
}

func TestAPI_ConfigEntries_TerminatingGateway(t *testing.T) {
	t.Parallel()
	c, s := makeClient(t)
	defer s.Stop()

	configEntries := c.ConfigEntries()

	terminating1 := &TerminatingGatewayConfigEntry{
		Kind: TerminatingGateway,
		Name: "foo",
	}

	terminating2 := &TerminatingGatewayConfigEntry{
		Kind: TerminatingGateway,
		Name: "bar",
	}

	// set it
	_, wm, err := configEntries.Set(terminating1, nil)
	require.NoError(t, err)
	require.NotNil(t, wm)
	require.NotEqual(t, 0, wm.RequestTime)

	// also set the second one
	_, wm, err = configEntries.Set(terminating2, nil)
	require.NoError(t, err)
	require.NotNil(t, wm)
	require.NotEqual(t, 0, wm.RequestTime)

	// get it
	entry, qm, err := configEntries.Get(TerminatingGateway, "foo", nil)
	require.NoError(t, err)
	require.NotNil(t, qm)
	require.NotEqual(t, 0, qm.RequestTime)

	// verify it
	readTerminating, ok := entry.(*TerminatingGatewayConfigEntry)
	require.True(t, ok)
	require.Equal(t, terminating1.Kind, readTerminating.Kind)
	require.Equal(t, terminating1.Name, readTerminating.Name)

	// update it
	terminating1.Services = []LinkedService{
		{
			Name:     "web",
			CAFile:   "/etc/web/ca.crt",
			CertFile: "/etc/web/client.crt",
			KeyFile:  "/etc/web/tls.key",
		},
	}

	// CAS fail
	written, _, err := configEntries.CAS(terminating1, 0, nil)
	require.NoError(t, err)
	require.False(t, written)

	// CAS success
	written, wm, err = configEntries.CAS(terminating1, readTerminating.ModifyIndex, nil)
	require.NoError(t, err)
	require.NotNil(t, wm)
	require.NotEqual(t, 0, wm.RequestTime)
	require.True(t, written)

	// re-setting should not yield an error
	_, wm, err = configEntries.Set(terminating1, nil)
	require.NoError(t, err)
	require.NotNil(t, wm)
	require.NotEqual(t, 0, wm.RequestTime)

	// web is associated with the other gateway, should get an error
	terminating2.Services = []LinkedService{
		{
			Name:     "*",
			CAFile:   "/etc/certs/ca.crt",
			CertFile: "/etc/certs/client.crt",
			KeyFile:  "/etc/certs/tls.key",
		},
		{
			Name: "web",
		},
	}
	_, wm, err = configEntries.Set(terminating2, nil)
	require.Error(t, err, "service \"web\" is associated with a different gateway")
	require.Nil(t, wm)

	// try again without web
	terminating2.Services = []LinkedService{
		{
			Name:     "*",
			CAFile:   "/etc/certs/ca.crt",
			CertFile: "/etc/certs/client.crt",
			KeyFile:  "/etc/certs/tls.key",
		},
	}
	_, wm, err = configEntries.Set(terminating2, nil)
	require.NoError(t, err)
	require.NotNil(t, wm)
	require.NotEqual(t, 0, wm.RequestTime)

	// list them
	entries, qm, err := configEntries.List(TerminatingGateway, nil)
	require.NoError(t, err)
	require.NotNil(t, qm)
	require.NotEqual(t, 0, qm.RequestTime)
	require.Len(t, entries, 2)

	for _, entry = range entries {
		switch entry.GetName() {
		case "foo":
			// this also verifies that the update value was persisted and
			// the updated values are seen
			readTerminating, ok = entry.(*TerminatingGatewayConfigEntry)
			require.True(t, ok)
			require.Equal(t, terminating1.Kind, readTerminating.Kind)
			require.Equal(t, terminating1.Name, readTerminating.Name)
			require.Equal(t, terminating1.Services, readTerminating.Services)
		case "bar":
			readTerminating, ok = entry.(*TerminatingGatewayConfigEntry)
			require.True(t, ok)
			require.Equal(t, terminating2.Kind, readTerminating.Kind)
			require.Equal(t, terminating2.Name, readTerminating.Name)
			require.Equal(t, terminating2.Services, readTerminating.Services)
		}
	}

	// delete it
	wm, err = configEntries.Delete(TerminatingGateway, "foo", nil)
	require.NoError(t, err)
	require.NotNil(t, wm)
	require.NotEqual(t, 0, wm.RequestTime)

	// verify deletion
	entry, qm, err = configEntries.Get(TerminatingGateway, "foo", nil)
	require.Error(t, err)
}
