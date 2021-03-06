package main

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/datasource"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"

	"github.com/prometheus/prometheus/promql/parser"
)

// newDatasource returns datasource.ServeOpts.
func newDatasource() datasource.ServeOpts {
	// creates a instance manager for your plugin. The function passed
	// into `NewInstanceManger` is called when the instance is created
	// for the first time or when a datasource configuration changed.
	im := datasource.NewInstanceManager(newDataSourceInstance)
	ds := &SampleDatasource{
		im: im,
	}

	return datasource.ServeOpts{
		QueryDataHandler:   ds,
		CheckHealthHandler: ds,
	}
}

// SampleDatasource is an example datasource used to scaffold
// new datasource plugins with an backend.
type SampleDatasource struct {
	// The instance manager can help with lifecycle management
	// of datasource instances in plugins. It's not a requirements
	// but a best practice that we recommend that you follow.
	im instancemgmt.InstanceManager
}

// QueryData handles multiple queries and returns multiple responses.
// req contains the queries []DataQuery (where each query contains RefID as a unique identifer).
// The QueryDataResponse contains a map of RefID to the response for each query, and each response
// contains Frames ([]*Frame).
func (td *SampleDatasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	log.DefaultLogger.Info("QueryData", "request", req)

	// create response struct
	response := backend.NewQueryDataResponse()

	// loop over queries and execute them individually.
	for _, q := range req.Queries {
		res := td.query(ctx, q)

		// save the response in a hashmap
		// based on with RefID as identifier
		response.Responses[q.RefID] = res
	}

	return response, nil
}

type queryModel struct {
	Format string `json:"format"`
}

type visitor parser.Visitor

type visited struct {
	visitor visitor
}

func (vis visited) Visit(node parser.Node, path []parser.Node) (w parser.Visitor, err error) {
	log.DefaultLogger.Info(fmt.Sprintf("%T %#v", node, node))
	switch node.(type) {
	case *parser.BinaryExpr:
		log.DefaultLogger.Info("BINARY EXPR")
	case *parser.AggregateExpr:
		log.DefaultLogger.Info("AGGREGATE EXPR")
	case *parser.NumberLiteral:
		log.DefaultLogger.Info("NUMBER LITERAL")
	case *parser.StringLiteral:
		log.DefaultLogger.Info("STRING LITERAL")
	case *parser.EvalStmt:
		log.DefaultLogger.Info("EVAL STMT")
	case *parser.Call:
		log.DefaultLogger.Info("EVAL STMT")
	case *parser.MatrixSelector:
		log.DefaultLogger.Info("MATRIX SELECTOR")
	case *parser.ParenExpr:
		log.DefaultLogger.Info("PAREN EXPR")
	case *parser.SubqueryExpr:
		log.DefaultLogger.Info("SUBQUERY EXPR")
	case *parser.UnaryExpr:
		log.DefaultLogger.Info("UNARY EXPR")
	case *parser.VectorSelector:
		log.DefaultLogger.Info("VECTOR SELECTOR")
	}
	return vis, err
}

func (td *SampleDatasource) query(ctx context.Context, query backend.DataQuery) backend.DataResponse {
	// Unmarshal the json into our queryModel
	// var qm queryModel

	log.DefaultLogger.Info(string(query.JSON))

	response := backend.DataResponse{}

	var parsed, parseError = parser.ParseExpr("a_metric_name{label='value', label2='value2'}")

	var visited visited

	parser.Walk(visited, parsed, []parser.Node{})

	if parseError != nil {
		log.DefaultLogger.Error("parsing borked %s", parseError)
		// response.Error = error
		// return response
		return response
	}

	log.DefaultLogger.Warn(parsed.String())
	log.DefaultLogger.Info(fmt.Sprintf("%T %#v", parsed, parsed))
	// implement the Visitor interface, and pass it to Walk
	// The visitor will implement the type switch in order to serialize to JSON
	// log.DefaultLogger.Warn(string(parsed.Type()))
	// switch v := parsed.(type) {
	// case *parser.VectorSelector:
	//     // do something with v, which has type *parser.VectorSelector here
	// case ...:
	// }
	frame := data.NewFrame("response")

	// frame.Fields = append(frame.Fields,
	// 	data.NewField(("AST"), parsed.End().IsValid(), []string),
	// )

	response.Frames = append(response.Frames, frame)
	return response

	// response.Error = json.Unmarshal(query.JSON, &qm)
	// if response.Error != nil {
	// 	return response
	// }

	// // Log a warning if `Format` is empty.
	// if qm.Format == "" {
	// 	log.DefaultLogger.Warn("format is empty. defaulting to time series")
	// }

	// // create data frame response
	// frame := data.NewFrame("response")

	// // add the time dimension
	// frame.Fields = append(frame.Fields,
	// 	data.NewField("time", nil, []time.Time{query.TimeRange.From, query.TimeRange.To}),
	// )

	// // add values
	// frame.Fields = append(frame.Fields,
	// 	data.NewField("values", nil, []int64{10, 20}),
	// )

	// // add the frames to the response
	// response.Frames = append(response.Frames, frame)

	// return response
}

// CheckHealth handles health checks sent from Grafana to the plugin.
// The main use case for these health checks is the test button on the
// datasource configuration page which allows users to verify that
// a datasource is working as expected.
func (td *SampleDatasource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	var status = backend.HealthStatusOk
	var message = "Data source is working"

	if rand.Int()%2 == 0 {
		status = backend.HealthStatusError
		message = "randomized error"
	}

	return &backend.CheckHealthResult{
		Status:  status,
		Message: message,
	}, nil
}

type instanceSettings struct {
	httpClient *http.Client
}

func newDataSourceInstance(setting backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	return &instanceSettings{
		httpClient: &http.Client{},
	}, nil
}

func (s *instanceSettings) Dispose() {
	// Called before creatinga a new instance to allow plugin authors
	// to cleanup.
}
