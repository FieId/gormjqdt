package gormjqdt

import (
	"log"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

func ParsingRequest(request RequestString) *ParsedRequest {
	// Initialize map of string
	var parsed map[string][]string

	switch true {
	// TODO: Add feature to support Json Body
	case (strings.Contains(string(request), ",") && strings.Contains(string(request), "{")):
		// var unboxit map[string]interface{}
		// json.Unmarshal([]byte(request), &parsed)
		// log.Printf("unboxit: %v", unboxit)
		// parsed = unboxit

	case strings.Contains(string(request), "&"):
		unboxit, err := url.ParseQuery(string(request))
		if err == nil {
			parsed = unboxit
		}

	default:
		return nil
	}

	return request.parsingParameters(parsed)
}

func (r RequestString) parsingParameters(parsed map[string][]string) *ParsedRequest {
	// Construct the struct to collect the parsed request
	parsedReq := &ParsedRequest{}
	parsedReq.Columns = make(map[string]interface{})
	parsedReq.Orders = make(map[string]interface{})
	parsedReq.SpesificParams = make(map[int]map[string]interface{})
	parsedReq.SpesificParamKeySlices = make(map[string]int)

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
