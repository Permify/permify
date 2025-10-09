# Permify Playground

This directory contains a Permify playground. You can check out running version of [Playground](https://play.permify.co).
Playground works with [webassembly](https://webassembly.org). You can find the source code of playground in [/pkg/development/wasm](/pkg/development/wasm)

## Running Playground Locally

To get started, install [wasm-pack](https://rustwasm.github.io/wasm-pack/installer/) and build the `play.wasm` file. Run this command from the root directory:

<!-- Build step -->
```sh
make wasm-build  # Compile WebAssembly module
```  
<!-- Start from playground directory -->
After building, start the playground from the playground directory:
<!-- Development server -->
```sh  
yarn start  # Launch development server
```  
Alternatively, run it from the root directory:
<!-- Alternative start method -->
```sh  
make serve-playground  # Alternative server start
```  
<!-- End local setup -->
<!-- Script documentation section -->
## Available Scripts

In the playground directory, you can run:

### `yarn start`

Runs the app in the development mode.\
Open [http://localhost:3000](http://localhost:3000) to view it in the browser.

The page will reload if you make edits.\
You will also see any lint errors in the console.

### `yarn test`

Launches the test runner in the interactive watch mode.\
See the section about [running tests](https://facebook.github.io/create-react-app/docs/running-tests) for more information.

### `yarn build`

Builds the app for production to the `build` folder.\
It correctly bundles React in production mode and optimizes the build for the best performance.

The build is minified and the filenames include the hashes.\
Your app is ready to be deployed!

See the section about [deployment](https://facebook.github.io/create-react-app/docs/deployment) for more information.
