package osquerytables

import (
	"context"

	osquerygo "github.com/osquery/osquery-go"
	"github.com/osquery/osquery-go/gen/osquery"
	"github.com/osquery/osquery-go/plugin/table"
)

// The pinned osquery-go table API cannot set column options. osquery uses bit 1
// for the INDEX option in the extension route schema.
const indexColumnOption = "1"

type indexedTablePlugin struct {
	plugin         *table.Plugin
	indexedColumns map[string]struct{}
}

func newIndexedTablePlugin(
	name string,
	columns []table.ColumnDefinition,
	generate table.GenerateFunc,
	indexedColumns ...string,
) osquerygo.OsqueryPlugin {
	indexed := make(map[string]struct{}, len(indexedColumns))
	for _, column := range indexedColumns {
		indexed[column] = struct{}{}
	}

	return &indexedTablePlugin{
		plugin:         table.NewPlugin(name, columns, generate),
		indexedColumns: indexed,
	}
}

func (plugin *indexedTablePlugin) Name() string {
	return plugin.plugin.Name()
}

func (plugin *indexedTablePlugin) RegistryName() string {
	return plugin.plugin.RegistryName()
}

func (plugin *indexedTablePlugin) Routes() osquery.ExtensionPluginResponse {
	return plugin.withIndexOptions(plugin.plugin.Routes())
}

func (plugin *indexedTablePlugin) Ping() osquery.ExtensionStatus {
	return plugin.plugin.Ping()
}

func (plugin *indexedTablePlugin) Call(
	ctx context.Context,
	request osquery.ExtensionPluginRequest,
) osquery.ExtensionResponse {
	response := plugin.plugin.Call(ctx, request)
	if request["action"] == "columns" {
		response.Response = plugin.withIndexOptions(response.Response)
	}
	return response
}

func (plugin *indexedTablePlugin) Shutdown() {
	plugin.plugin.Shutdown()
}

func (plugin *indexedTablePlugin) withIndexOptions(
	routes osquery.ExtensionPluginResponse,
) osquery.ExtensionPluginResponse {
	for _, route := range routes {
		if _, indexed := plugin.indexedColumns[route["name"]]; indexed {
			route["op"] = indexColumnOption
		}
	}
	return routes
}
