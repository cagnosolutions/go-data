This is going to *hopefully* contain the final revisions of the `dbms` package \
with all the refactoring and revisions ironed out.
---
This (engine) is the successor to [go-data/pkg/dbms](https://github.com/cagnosolutions/go-data/pkg/dbms),
which is the successor to [go-data/pkg/pager](https://github.com/cagnosolutions/go-data/pkg/pager)

<script src="https://cdn.jsdelivr.net/npm/mathjax@3/es5/tex-mml-chtml.js"></script>

```go
// Open a storage engine instance
db, err := OpenEngine("my/db")
if err != nil {
panic(err)
}

// Don't forget to close!
defer func (db *Engine){
err := db.Close()
if err != nil {
panic(err)
}
}(db)

// Get a namespace (create if not exist)
ns, err := db.Namespace("users")
if err != nil {
panic(err)
}

// Insert an entry
id, err := ns.Add(userData)
if err != nil {
panic(err)
}

// Update an entry
err := ns.Put(id, userData)
if err != nil {
panic(err)
}

// Return an entry
userData, err := ns.Get(id)
if err != nil {
panic(err)
}

// Commit whatever you did
err := ns.Commit()
if err != nil {
panic(err)
}
```

```go
// Create a namespace
ns, err := db.CreateNamespace("users")
if err != nil  {
panic(err)
}
```

```go
// Drop a namespace
err := db.DropNamespace("users")
if err != nil {
panic(err)
}
```

```go
// Use a namespace
ns, err := db.Namespace("users")
if err != nil {
panic(err)
}

// Marshal user into bytes
userData, err := json.Marshal(u)
if err != nil {
panic(err)
}

// Insert a new user
id, err := ns.Insert(userData)
if err != nil {
panic(err)
}

// Update an existing user
err := ns.Update(id, userData)
if err != nil {
panic(err)
}

// Get by id
userData, err := ns.Find(id)
if err != nil {
panic(err)
}
```