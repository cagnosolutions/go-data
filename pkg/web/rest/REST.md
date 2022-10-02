# RESTful APIs

The term REST was suggested by Roy Fielding in his Ph.D. dissertation in the year 2000. \
It stands for ***R***epresentational ***S***tate ***T***ransfer and is described as:

*"REST emphasizes scalability of component interactions, generality of interfaces, \
independent deployment of components, and intermediary components to reduce interaction \
latency, enforce security and encapsulate legacy systems."*

Having an API that conforms to the ***REST*** principles is what makes it ***REST***ful.
___

# Contents

- [URI Format](#uri-format)
- [Path Design](#path-design)
    - [Documents](#documents)
    - [Collections](#collections)
    - [Stores](#stores)
    - [Controllers](#controllers)
- [CRUD Function Names](#crud-function-names)
- [HTTP Verbs](#http-verbs)
    - [GET](#method-get)
    - [POST](#method-post)
    - [PUT](#method-put)
    - [PATCH](#method-patch)
    - [DELETE](#method-delete)
    - [HEAD](#method-head)
    - [OPTIONS](#method-options)
- [Response Codes](#response-codes)
    - [2xx Success](#2xx-success)
        - [200 OK](#200-ok)
        - [201 Created](#201-created)
        - [204 No Content](#204-no-content)
    - [3xx Redirection](#3xx-redirection)
        - [301 Moved Permanently](#301-moved-permanently)
        - [304 Not Modified](#304-not-modified)
    - [4xx Client Error](#4xx-client-error)
        - [400 Bad Request](#400-bad-request)
        - [401 Unauthorized](#401-unauthorized)
        - [403 Forbidden](#403-forbidden)
        - [404 Not Found](#404-not-found)
        - [405 Method Not Allowed](#405-method-not-allowed)
        - [408 Request Timeout](#408-request-timeout)
    - [5xx ServerError](#5xx-server-error)
        - [500 Internal Server Error](#500-internal-server-error)
        - [503 Service Unavailable](#503-service-unavailable)

# URI Format

RFC 3986, which was published in 2005 https://www.ietf.org/rfc/rfc3986.txt, defines the \
format that makes valid URIs:

```
URI = scheme "://" authority "/" path [ "?" query] ["#" fragment"]
URI = http://myserver.com/mypath?query=1#document
```

We will use the path element in order to locate an endpoint that is running on our server. \
In a REST endpoint, this can contain parameters as well as a document location. The query \
string is equally important, as you will use this to pass parameters such as page number or \
ordering to control the data that is returned.

Some general rules for URI formatting:

- A forward slash `/` is used to indicate a hierarchical relationship between resources
- A trailing forward slash `/` should not be included in URIs
- Hyphens `-` should be used to improve readability
- Underscores `_` should not be used in URIs
- Lowercase letters are preferred as case sensitivity is a differentiator in the path part \
  of a URI

The concept behind many of the rules is that a URI should be easy to read and to construct. \
It should also be consistent in the way that it is built, so you should follow the same \
taxonomy for all the endpoints in your API.
___

# Path Design

Paths are broken into [*documents*](#documents), [*collections*](#collections),
[*stores*](#stores), and [*controllers*](#controllers). For a quick overview:

- A ***document*** is a resource pointing to a single object, like a row in a database.
- A ***collection*** is a set of documents, very much like a table in a database.
- A ***store*** is a set of client managed collections, very much like a database.
- A ***controller*** is a functional procedure that is not a typical CRUD operation.

### Documents

A document is a resource pointing to a single object, similar to a row in a database. It has \
the ability to have child resources that may be both sub-documents or collections.

For example:

```http
GET /cats/1 -> Single document for cat 1
GET /cats/1/kittens -> All kittens belonging to cat 1
GET /cats/1/kittens/1 -> Kitten 1 for cat 1
```

### Collections

A collection is a directory of resources typically broken by parameters to access an individual \
document.

For example:

```http
GET /cats -> All cats in the collection
GET /cats/1 -> Single document for a cat 1
```

When defining a collection, we should *always use a plural noun* such as cats or people for the\
collection name.

### Stores

A store is a client-managed resource repository, it allows the client to add, retrieve, and \
delete resources. Unlike a collection, a store will never generate a new URI it will use the \
one specified by the client. Take a look at the following example that would add a new cat to\
our store:

```http
PUT /cats/2
```

This would add a new cat to the store with an ID of 2, if we had posted the new cat omitting \
the ID to a collection the response would need to include a reference to the newly defined\
document, so we could later interact with it. Like controllers, we should use a plural noun \
for store names.

### Controllers

A controller resource is like a procedure, this is typically used when a resource cannot be \
mapped to standard CRUD (create, retrieve, update, and delete) functions.

The names for controllers appear as the last segment in a URI path with no child resources. If \
the controller requires parameters, these would typically be included in the query string:

```http
POST /cats/1/feed -> Feed cat 1
POST /cats/1/feed?food=fish ->Feed cat 1 a fish
```

When defining a controller name we should always use a verb. A verb is a word that indicates an\
action or a state of being, such as feed or send.
___

# CRUD Function Names

When designing great REST URIs we never use a CRUD function name as part of the URI, instead \
we use a HTTP verb.

For example:

```http
DELETE /cats/1234
```

We do not include the verb in the name of the method as this is specified by the HTTP verb, \
the following URIs would be considered an anti-pattern:

```http
GET /deleteCat/1234
DELETE /deleteCat/1234
POST /cats/1234/delete
```

When we look at HTTP verbs in the next section this will make more sense.

# HTTP Verbs

- [GET](#method-get)
- [POST](#method-post)
- [PUT](#method-put)
- [PATCH](#method-patch)
- [DELETE](#method-delete)
- [HEAD](#method-head)
- [OPTIONS](#method-options)

Each of these methods has a well-defined semantic within the context of our REST API and the \
correct implementation will help your user understand your intention.
___

### Method GET

The GET method is used to retrieve a resource and should never be used to mutate an operation, \
such as updating a record. Typically, a body is not passed with a GET request; however, it is \
not an invalid HTTP request to do so.

***Request:***

```http
GET /v1/cats HTTP/1.1
```

***Response:***

```
HTTP/1.1 200 OK
Content-Type: application/json
Content-Length: xxxx

{"name": "Fat Freddie's Cat", "weight": 15}
```

___

### Method POST

The POST method is used to create a new resource in a collection or to execute a controller. It\
is typically a non-idempotent action, in that multiple posts to create an element in a collection\
that will create multiple elements not updated after the first call.

The POST method is always used when calling controllers as the actions of this is considered \
non-idempotent.

***Request:***

```http
POST /v1/cats HTTP/1.1
Content-Type: application/json
Content-Length: xxxx

{"name": "Felix", "weight": 5}
```

***Response:***

```
HTTP/1.1 201 Created
Content-Type: application/json
Content-Length: 0
Location: /v1/cats/12343
```

___

### Method PUT

The PUT method is used to update a mutable resource and must always include the resource locator.\
The PUT method calls are also idempotent in that multiple requests will not mutate the resource\
to a different state than the first call.

***Request:***

```http
PUT /v1/cats HTTP/1.1
Content-Type: application/json
Content-Length: xxxx

{"name": "Thomas", "weight": 7 }
```

***Response:***

```
HTTP/1.1 201 Created
Content-Type: application/json
Content-Length: 0
```

___

### Method PATCH

The PATCH verb is used to perform a partial update, for example, if we only wanted to update \
the name of our cat we could make a PATCH request only containing the details that we would \
like to change.

***Request:***

```http
PATCH /v1/cats/12343 HTTP/1.1
Content-Type: application/json
Content-Length: xxxx

{"weight": 9}
```

***Response:***

```
HTTP/1.1 204 No Body
Content-Type: application/json
Content-Length: 0
```

In my experience PATCH updates are rarely used, the general convention is to use a PUT and \
to update the whole object, this not only makes the code easier to write but also makes an \
API which is simpler to understand.
___

### Method DELETE

The DELETE verb is used when we want to remove a resource, generally we would pass the ID of \
the resource as part of the path rather than in the body of the request. This way, we have a \
consistent method for updating, deleting, and retrieving a document.

***Request:***

```http
DELETE /v1/cats/12343 HTTP/1.1
Content-Type: application/json
Content-Length: 0
```

***Response:***

```
HTTP/1.1 204 No Body
Content-Type: application/json
Content-Length: 0
```

___

### Method HEAD

A client would use the HEAD verb when they would like to retrieve the headers for a resource \
without the body. The HEAD verb is typically used in place of a GET verb when a client only \
wants to check if a resource exists or to read the metadata.

***Request:***

```http
HEAD /v1/cats/12343 HTTP/1.1
Content-Type: application/json
Content-Length: 0
```

***Response:***

```
HTTP/1.1 200 OK
Content-Type: application/json
Last-Modified: Wed, 25 Feb 2004 22:37:23 GMT
Content-Length: 45
```

___

### Method OPTIONS

The OPTIONS verb is used when a client would like to retrieve the possible interactions for a \
resource. Typically, the server will return an Allow header, which will include the HTTP verbs \
that can be used with this resource.

***Request:***

```http
OPTIONS /v1/cats/12343 HTTP/1.1
Content-Length: 0
```

***Response:***

```
HTTP/1.1 200 OK
Content-Length: 0
Allow: GET, PUT, DELETE
```

___

# Response Codes

When writing a great API, we should use HTTP status codes to indicate to the client the success \
or failure of the request. We will not be taking a comprehensive look, but we will look at the \
status codes that you will want your microservice to return.
___
---

## 2XX Success

The 2XX status codes indicate that the clients request has been successfully received and understood.
___

### 200 OK

This is a generic response code indicating that the request has succeeded. The response accompanying \
this code is generally:

- GET: An, an entity corresponding to the requested resource
- HEAD: The, the header fields corresponding to the requested resource without the message body
- POST: An, an entity describing or containing the result of the action

___

### 201 Created

The created response is sent when a request succeeds and the result is that a new entity has been \
created. Along with the response it is common that the API will return a Location header with the\
location of the newly created entity:

```
201 Created
Location: https://api.kittens.com/v1/kittens/123dfdf111
```

It is optional to return an object body with this response type.
___

### 204 No Content

This status informs the client that the request has been successfully processed; however, there\
will be no message body with the response. For example, if the user makes a DELETE request to the\
collection then the response may return a 204 status.
___
---

## 3XX Redirection

The 3xx indicate class of status codes indicates that the client must take additional action to \
complete the request. Many of these status codes are used by CDNs and other content redirection \
techniques, however, code 304 can exceptionally useful when designing our APIs to provide semantic\
feedback to the client.
___

### 301 Moved Permanently

This tells the client that the resource they have requested has been permanently moved to a \
different location. Whilst this is traditionally used to redirect a page or resource from a web \
server it can also be useful to us when we are building our APIs. In the instance that we rename \
a collection we could use a 301 redirect to send the client to the correct location. This however \
should be used as an exception rather than the norm. Some clients do not implicitly follow 301 \
redirect and implementing this capability adds additional complexity for your consumers.
___

### 304 Not Modified

This response is generally used by a CDN or caching server and is set to indicate that the \
response has not been modified since the last call to the API. This is designed to save bandwidth\
and the request will not return a body, but will return a ContentLocation and Expires header.
___
---

## 4XX Client Error

In the instance of an error caused by a client, not the server, the server will return a 4xx \
response and will always return an entity that gives further details on the error.

### 400 Bad Request

This response indicates that the request could not be understood by the client due to a malformed\
request or due to a failure of domain validation (missing data, or an operation that would cause \
invalid state).
___

### 401 Unauthorized

This indicates that the request requires user authentication and will include a WWW-Authenticate \
header containing a challenge applicable to the requested resource. If the user has included the \
required credentials in the WWW-Authenticate header, then the response should include an error \
object that may contain relevant diagnostic information.
___

### 403 Forbidden

The server has understood the request, but is refusing to fulfill it. This could be due to \
incorrect access level to a resource not that the user is not authenticated.

If the server does not wish to make the fact that a request is not able to access a resource \
due to access level public, then it is permissible to return a 404 Not found status instead \
of this response.
___

### 404 Not Found

This response indicates that the server has not found anything matching the requested URI. No \
indication is given of whether the condition is temporary or permanent.

It is permissible for the client to make multiple requests to this endpoint as the state may not \
be permanent.
___

### 405 Method Not Allowed

The method specified in the request is not allowed for the resource indicated by the URI. This may\
be when the client attempts to mutate a collection by sending a POST, PUT, or PATCH to a \
collection that only serves retrieval of documents.
___

### 408 Request Timeout

The client did not produce a request within the time that the server is prepared to wait. The \
client may repeat the request without modification at a later time.
___
---

## 5XX Server Error

Response status codes within the 500 range indicate that something has gone "Bang", the server \
knows this and is sorry for the situation.

The RFC advises that an error entity should be returned to the response explaining whether this \
is permanent or temporary and containing an explanation of the error. When we look at our chapter \
on security we will look at the recommendation about not giving too much information away in error \
messages as this state may have been engineered by a user in the attempt to compromise your system \
and by returning things such as a stack trace or other internal information with a 5xx error can \
actually help to compromise your system. With this in mind it is currently common that a 500 error \
will just return something very generic.
___

### 500 Internal Server Error

A generic error message indicating that something did not go quite as planned.
___

### 503 Service Unavailable

The server is currently unavailable due to temporary overloading or maintenance. There is a rather\
useful pattern that you can implement to avoid cascading failure in the instance of a malfunction\
in which the microservice will monitor its internal state and in the case of failure or overloading\
will refuse to accept the request and immediately signal this to the client. We will look at this \
pattern more in chapter xx; however, this instance is probably where you will be wanting to return \
a 503 status code. This could also be used as part of your health checks.
___