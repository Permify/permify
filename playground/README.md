# Permify Playground

This directory contains a Permify playground. You can check out running version of [Playground](https://play.permify.co).
Playground works with [webassembly](https://webassembly.org). You can find the source code of playground in [/pkg/development/wasm](/pkg/development/wasm)

## How to run playground locally
First, you need to install [wasm-pack](https://rustwasm.github.io/wasm-pack/installer/). Then, you will need to build the `play.wasm` file. You can do it by running the following command in the root directory:

```bash
make wasm-build
```

Then, you can run the playground by running the following command in the playground directory:

```bash
yarn start
```
or the following command in the root directory:

```bash
make serve-playground
```

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
