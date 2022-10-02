package rest

func IsValidCode(code int) bool {
	if Is2xxCode(code) || Is3xxCode(code) || Is4xxCode(code) || Is5xxCode(code) {
		_, found := httpCodes[code]
		return found
	}
	return false
}

func Is2xxCode(code int) bool {
	return (code / 200) == 1
}

func Is3xxCode(code int) bool {
	return (code / 300) == 1
}

func Is4xxCode(code int) bool {
	return (code / 400) == 1
}

func Is5xxCode(code int) bool {
	return (code / 500) == 1
}

var httpCodes = map[int][2]string{
	//
	// 2xx status codes (success)
	200: {
		"OK",
		"Indicates that the request has succeeded.",
	},
	201: {
		"Created",
		"Indicates that the request has succeeded and a new resource has been created as a result.",
	},
	202: {
		"Accepted",
		"Indicates that the request has been received but not completed yet. It is typically used in log running " +
			"requests and batch processing.",
	},
	204: {
		"No Content",
		"The server has fulfilled the request but does not need to return a response body. The server may return " +
			"the updated meta information.",
	},
	//
	// 3xx status codes (redirection)
	301: {
		"Moved Permanently",
		"The URL of the requested resource has been changed permanently. The new URL is given by the Location header " +
			"field in the response. This response is cacheable unless indicated otherwise.",
	},
	302: {
		"Found",
		"The URL of the requested resource has been changed temporarily. The new URL is given by the Location field " +
			"in the response. This response is only cacheable if indicated by a Cache-Control or Expires header field.",
	},
	303: {
		"See Other",
		"The response can be found under a different URI and SHOULD be retrieved using a GET method on that resource.",
	},
	304: {
		"Not Modified",
		"Indicates the client that the response has not been modified, " +
			"so the client can continue to use the same cached version of the response.",
	},
	307: {
		"Temporary Redirect",
		"Indicates the client to get the requested resource at another URI with same method that was used in the " +
			"prior request. It is similar to 302 Found with one exception that the same HTTP method will be used " +
			"that was used in the prior request.",
	},
	//
	// 4xx status codes (client error)
	400: {
		"Bad Request",
		"The request could not be understood by the server due to incorrect syntax. " +
			"The client SHOULD NOT repeat the request without modifications.",
	},
	401: {
		"Unauthorized",
		"Indicates that the request requires user authentication information. " +
			"The client MAY repeat the request with a suitable Authorization header field",
	},
	403: {
		"Forbidden",
		"Unauthorized request. The client does not have access rights to the content. Unlike 401, " +
			"the client’s identity is known to the server.",
	},
	404: {
		"Not Found",
		"The server can not find the requested resource.",
	},
	405: {
		"Method Not Allowed",
		"The request HTTP method is known by the server but has been disabled and cannot be used for that resource.",
		// reply with `Allow: GET, POST`, etc in the header
	},
	406: {
		"Not Acceptable",
		"The server doesn’t find any content that conforms to the criteria given by the user agent in the Accept" +
			" header sent in the request.",
	},
	408: {
		"Request Timeout",
		"Indicates that the server did not receive a complete request from the client within the server’s allotted" +
			" timeout period.",
	},
	409: {
		"Conflict",
		"The request could not be completed due to a conflict with the current state of the resource.",
	},
	410: {
		"Gone",
		"The requested resource is no longer available at the server.",
	},
	412: {
		"Precondition Failed",
		"The client has indicated preconditions in its headers which the server does not meet.",
	},
	415: {
		"Unsupported Media Type",
		"The media-type in Content-type of the request is not supported by the server.",
	},
	429: {
		"Too Many Requests",
		"The user has sent too many requests in a given amount of time (“rate limiting”).",
	},
	//
	// 5xx status codes (server error)
	500: {
		"Internal Server Error",
		"The server encountered an unexpected condition that prevented it from fulfilling the request.",
	},
	501: {
		"Not Implemented",
		"The HTTP method is not supported by the server and cannot be handled.",
	},
	503: {
		"Service Unavailable",
		"The server is not ready to handle the request.",
	},
}
