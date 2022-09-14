# poptape-categories
Categories microservice written in Go

This microservice creates/reads/updates/deletes category data in a 
Postgres database. Categories are used to supply field data for items and 
building forms in the react client.

### API routes

```
/categories [GET] (Unauthenticated)

Returns every category
Expected normal return codes: [200, 404]


/categories/<cat_id> [GET] (Unauthenticated)

Returns all data associated with a particular category
Expected normal return codes: [200, 404]


/categories/<cat_id>/children [GET] (Unauthenticated)

Returns all data associated with a particular and any
associated child categories 
Expected normal return codes: [200, 404]


/categories [POST] (Authenticated)

Create a category for the authenticated user.
Expected normal return codes: [201, 401]


/categories/<cat_id> [DELETE] (Authenticated)

Deletes a single category. Only works if category has no children.
Expected return codes: [401, 410]


/categories/top [GET] (Unauthenticated)

Returns all top level categories.
Expected return codes: [200]


/categories/second [GET] (Unauthenticated)

Returns all second level categories.
Expected return codes: [200]


/categories/third [GET] (Unauthenticated)

Returns all third level categories.
Expected return codes: [200]



```

### To Do:
* All of it :)
* Write tests
* Validate input
* Dockerize
* Documentation
