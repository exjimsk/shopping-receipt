# shopping-receipt

  `src/app/main.go` contains the code to kick up the http server.  

  The program reads the content of `_states.csv` and `_items.csv` on start. It contains information like item prices, state taxes, names, tax exempted item categories.

  The web app listens on 4 endpoints:
  - "/": serves the web document
  - "/states": serves json of states
  - "/items": serves json of items
  - "/total": responds to a POST request that contains a cart json with a receipt json

  The UI is a simple jQuery app using flexbox layout.

