package gormjqdt

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

func ParsingRequest(request RequestString) *ParsedRequest {
	switch true {
	case (strings.Contains(string(request), ",") && strings.Contains(string(request), "{")):
		var unboxit map[string]interface{}
		err := json.Unmarshal([]byte(request), &unboxit)
		if err != nil {
			return nil
		}
		return parsinJsonRequest(unboxit)

	case strings.Contains(string(request), "&"):
		unboxit, _ := url.ParseQuery(string(request))
		return parsingUrlEncodedRequest(unboxit)
	}

	return nil
}

func parsingUrlEncodedRequest(parsed map[string][]string) *ParsedRequest {
	// Construct the struct to collect the parsed request
	parsedReq := &ParsedRequest{}
	parsedReq.Columns = make(map[string]interface{})
	parsedReq.Orders = make(map[string]interface{})
	parsedReq.SpesificParams = make(map[int]map[string]interface{})
	parsedReq.SpesificParamKeySlices = make(map[string]int)

	// If request has empty or not detected
	if len(parsed) <= 0 {
		parsedReq.Draw = 0
		parsedReq.Start = 0
		parsedReq.Length = 10
	}

	// Loop and proccess
	var i int
	for key := range parsed {
		switch true {
		// Draw
		case strings.Contains(key, "draw"):
			draw, err := strconv.Atoi(GetValFromSlice(parsed, key))
			if err != nil {
				draw = 0
			}

			parsedReq.Draw = draw

		// Start
		case strings.Contains(key, "start"):
			start, err := strconv.Atoi(GetValFromSlice(parsed, key))
			if err != nil || start < 0 {
				start = 0
			}

			parsedReq.Start = start

		// Length
		case strings.Contains(key, "length"):
			length, err := strconv.Atoi(GetValFromSlice(parsed, key))
			if err != nil || length < 0 {
				length = 10
			}

			parsedReq.Length = length

		// Global Search
		case strings.Contains(key, "search[value]"):
			parsedReq.GlobalSearch = GetValFromSlice(parsed, key)

		// Global Search Regex
		case strings.Contains(key, "search[regex]"):
			serachRegex, err := strconv.ParseBool(GetValFromSlice(parsed, key))
			if err != nil {
				serachRegex = false
			}

			parsedReq.GlobalSearchRegex = serachRegex

		// Columns
		case strings.Contains(key, "columns"):
			parsedReq.Columns[key] = GetValFromSlice(parsed, key)

		// Orders
		case strings.Contains(key, "order"):
			parsedReq.Orders[key] = GetValFromSlice(parsed, key)

		// Spesific Params
		default:
			i++

			// Regexp to replace non alphabet and _ character
			reg, err := regexp.Compile("[^a-zA-Z_]+")
			if err != nil {
				log.Fatalf("[request.go - SpesificParams] regex error: %v", err)
			}
			keyRegx := reg.ReplaceAllString(key, "")
			indexSpesificParamSlice := parsedReq.SpesificParamKeySlices[keyRegx]

			// If the value of params is array and the index defined in params key
			// I.e: status[1] = 'someValue', status[3] = 'anotherValue'
			if indexSpesificParamSlice <= 0 {
				parsedReq.SpesificParams[i] = make(map[string]interface{})
				parsedReq.SpesificParams[i]["key"] = keyRegx
				parsedReq.SpesificParams[i]["value"] = parsed[key]

				// Store the current param key to slice/array temp variable
				parsedReq.SpesificParamKeySlices[keyRegx] = i
			} else {
				switch v := parsedReq.SpesificParams[indexSpesificParamSlice]["value"].(type) {
				case []string:
					parsedReq.SpesificParams[indexSpesificParamSlice]["value"] = append(v, GetValFromSlice(parsed, key))
				}
			}
		}
	}

	// Return the parsed request struct
	return parsedReq
}

func parsinJsonRequest(parsed map[string]interface{}) *ParsedRequest {
	// Construct the struct to collect the parsed request
	parsedReq := &ParsedRequest{}
	parsedReq.Columns = make(map[string]interface{})
	parsedReq.Orders = make(map[string]interface{})
	parsedReq.SpesificParams = make(map[int]map[string]interface{})
	parsedReq.SpesificParamKeySlices = make(map[string]int)

	var i int
	for k, v := range parsed {
		switch true {
		// Draw
		case k == "draw":
			draw, err := strconv.Atoi(ConvertInJsonValToString(v))
			if err != nil {
				draw = 0
			}
			parsedReq.Draw = draw

		// Start
		case k == "start":
			start, err := strconv.Atoi(ConvertInJsonValToString(v))
			if err != nil || start < 0 {
				start = 0
			}
			parsedReq.Start = start

		// Length
		case k == "length":
			length, err := strconv.Atoi(ConvertInJsonValToString(v))
			if err != nil || length < 0 {
				length = 10
			}
			parsedReq.Length = length

		// Orders
		case k == "order":
			for ork, orv := range v.([]interface{}) {
				parsedReq.Orders[fmt.Sprintf("order[%v][column]", ork)] = ConvertInJsonValToString(orv, "column")
				parsedReq.Orders[fmt.Sprintf("order[%v][dir]", ork)] = ConvertInJsonValToString(orv, "dir")
			}

		// Columns
		case k == "columns":
			for cok, cov := range v.([]interface{}) {
				for nek, nev := range cov.(map[string]interface{}) {
					if fmt.Sprintf("%v", reflect.TypeOf(nev)) == "map[string]interface {}" {
						for nesk, nesv := range nev.(map[string]interface{}) {
							parsedReq.Columns[fmt.Sprintf("columns[%v][%v][%v]", cok, nek, nesk)] = ConvertInJsonValToString(nesv, nesk)
						}
					} else {
						parsedReq.Columns[fmt.Sprintf("columns[%v][%v]", cok, nek)] = ConvertInJsonValToString(nev, nek)
					}
				}
			}

		// Global Search
		case k == "search":
			parsedReq.GlobalSearch = ConvertInJsonValToString(v, "value")
			serachRegex, err := strconv.ParseBool(ConvertInJsonValToString(v, "regex"))
			if err != nil {
				serachRegex = false
			}
			parsedReq.GlobalSearchRegex = serachRegex

		// Spesific Params
		default:
			i++
			parsedReq.SpesificParams[i] = make(map[string]interface{})
			parsedReq.SpesificParams[i]["key"] = k
			parsedReq.SpesificParams[i]["value"] = v
		}
	}

	// Return the parsed request struct
	return parsedReq
}
