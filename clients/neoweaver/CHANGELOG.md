# Changelog

## [0.1.4](https://github.com/nkapatos/mindweaver/compare/neoweaver/v0.1.3...neoweaver/v0.1.4) (2025-12-20)


### Bug Fixes

* **ci:** fix path filtering and add manual neoweaver sync trigger ([#25](https://github.com/nkapatos/mindweaver/issues/25)) ([4e88290](https://github.com/nkapatos/mindweaver/commit/4e882902beba417730a126455ff3dcb3986386be))

## [0.1.3](https://github.com/nkapatos/mindweaver/compare/neoweaver/v0.1.2...neoweaver/v0.1.3) (2025-12-20)


### Maintenance

* **neoweaver:** remove old api.bak file from v3 ([#23](https://github.com/nkapatos/mindweaver/issues/23)) ([79e2c5d](https://github.com/nkapatos/mindweaver/commit/79e2c5d0234573bb85aaebbce5375c1ea0181fa1))

## [0.1.2](https://github.com/nkapatos/mindweaver/compare/neoweaver/v0.1.1...neoweaver/v0.1.2) (2025-12-20)


### Features

* **neoweaver:** add collections tree explorer with API integration ([3d73836](https://github.com/nkapatos/mindweaver/commit/3d73836b428411290c25cc436f1955d5a6ab3e59))
* **neoweaver:** add dedicated repository sync workflow ([59c3e02](https://github.com/nkapatos/mindweaver/commit/59c3e02a7835d8cff9af6c13eb2fccb89bb928f0))
* **neoweaver:** add etag conflict detection and placeholder diff view ([312ea7f](https://github.com/nkapatos/mindweaver/commit/312ea7fd53020176122f9719762fc31066f52266))
* **neoweaver:** add immediate title editing ([550a496](https://github.com/nkapatos/mindweaver/commit/550a4964d6c12de1ae75d7def77733d203511e89))
* **neoweaver:** add LSP, linting, and formatting configs ([b2b13ac](https://github.com/nkapatos/mindweaver/commit/b2b13ac00b473b616b493801b1b06b060dbd4970))
* **neoweaver:** add mini.doc documentation generation infrastructure ([d9e5c01](https://github.com/nkapatos/mindweaver/commit/d9e5c013b11ef9322d3f0396413eec11f8198882))
* **neoweaver:** add note opening from explorer with smart window management ([1e06103](https://github.com/nkapatos/mindweaver/commit/1e06103aa72a2c92775dee9b2c51052c89bb62c7))
* **neoweaver:** add nui.nvim dependency and explorer foundation ([e035c3c](https://github.com/nkapatos/mindweaver/commit/e035c3ceef45734e2d958964c99c8893c4598661))
* **neoweaver:** add server context to explorer and buffer statuslines ([874c7b5](https://github.com/nkapatos/mindweaver/commit/874c7b5bbe774ec41ff8ccf247d407194a3e4888))
* **neoweaver:** add title workflows and namespaced commands ([cc91ca5](https://github.com/nkapatos/mindweaver/commit/cc91ca536fe65e280c02167240d7c4abeb305c64))
* **neoweaver:** add tree keybindings and commands for collection CRUD operations ([6755d56](https://github.com/nkapatos/mindweaver/commit/6755d561204012fc8b21bdd6a6a1545e3012c056))
* **neoweaver:** enhance conflict resolution messaging ([8e850ac](https://github.com/nkapatos/mindweaver/commit/8e850ac93704e90f45e6e9ac873dd1429e7ddf4e))
* **neoweaver:** implement collections CRUD operations (create, update, delete) ([e86631d](https://github.com/nkapatos/mindweaver/commit/e86631d2be8c8728802d51035066d1323b4e7ec3))
* **neoweaver:** implement multi-strategy conflict resolution ([a700306](https://github.com/nkapatos/mindweaver/commit/a700306cecc20f7af8060fff08daf6baf5436377))
* **neoweaver:** use field masking in collections tree for minimal note data ([b2446cc](https://github.com/nkapatos/mindweaver/commit/b2446cc64553f1c10ae0fc37e58f8bd82b9cbe54))


### Bug Fixes

* **neoweaver:** handle buffer modifiable state during tree refresh and load ([83f8c67](https://github.com/nkapatos/mindweaver/commit/83f8c67bb4c91a22e60af30d18f560a5d79d39ac))


### Documentation

* **neoweaver:** add acknowledgements for nui.nvim and neo-tree ([481e5c6](https://github.com/nkapatos/mindweaver/commit/481e5c6654b4a7018ab7ed86fcb2459dd46b8f1f))
* **neoweaver:** add generated API documentation files ([39ca92b](https://github.com/nkapatos/mindweaver/commit/39ca92ba4c8f2cd88bd08894cda659df8622ec40))
* **neoweaver:** add tree architecture refactoring plan ([066bec6](https://github.com/nkapatos/mindweaver/commit/066bec6a0cc9bd81aec4635bd049d1272bd27019))
* **neoweaver:** enhance API documentation with detailed annotations ([64fa57e](https://github.com/nkapatos/mindweaver/commit/64fa57e07cd76bda3cb98a8aa1dcee6f62add10c))


### Maintenance

* tooling in mise and ignore gen, tmp dirs from git ([fe994c4](https://github.com/nkapatos/mindweaver/commit/fe994c4ff92ab6a39ff5c2625f4d63fc13a67c87))


### Refactoring

* **neoweaver:** adopt best practices structure with _internal/ pattern ([34bb521](https://github.com/nkapatos/mindweaver/commit/34bb521604bdb0a8df99057c45b550026493b8c2))
* **neoweaver:** change the respose error handler to check for Connect error codes respose ([2e0d974](https://github.com/nkapatos/mindweaver/commit/2e0d974263fc0f39245032e5efc617f08047375e))
* **neoweaver:** decouple tree from domain logic, make explorer orchestrate actions ([98baeaa](https://github.com/nkapatos/mindweaver/commit/98baeaa5ac57245add876c3f62067d05f48fb4ed))
* **neoweaver:** organize plugin following Neovim best practices ([3d331eb](https://github.com/nkapatos/mindweaver/commit/3d331eb396a5e32061d1e16fefc26d43e11ab251))
* **neoweaver:** simplify explorer buffer management ([c819ff0](https://github.com/nkapatos/mindweaver/commit/c819ff09885e1f8c31f9604eae3df0f9bd17ec6c))
* **neoweaver:** simplify Taskfile paths by removing redundant NEOWEAVER_DIR ([f50396c](https://github.com/nkapatos/mindweaver/commit/f50396c21ed363691aec1085d69472487e09bba7))
* **neoweaver:** use helper for buffer writes, rely on NuiTree for render modifiable handling ([24267ec](https://github.com/nkapatos/mindweaver/commit/24267ec427faee88ede40b6b3b5ab0162236516b))

## [0.1.1](https://github.com/nkapatos/mindweaver/compare/neoweaver/v0.1.0...neoweaver/v0.1.1) (2025-12-16)


### Features

* establish component-specific structure ([1af0de2](https://github.com/nkapatos/mindweaver/commit/1af0de278fa3feef5bafd9c11b50081b56e001e1))
* implement generic buffer manager for neoweaver ([b4456c9](https://github.com/nkapatos/mindweaver/commit/b4456c96992f1af11e41482cd9b9aea0575cedb6))
* implement note open and save using buffer manager ([343bd49](https://github.com/nkapatos/mindweaver/commit/343bd49839e33ceaf020c6eef0b3043630abb35a))
* implement server-first note creation in neoweaver ([bf0df2d](https://github.com/nkapatos/mindweaver/commit/bf0df2dc869d90fb5e743538ffff6ded7d10ccc4))
* neoweaver v3 migration - API layer and minimal notes ([a4e9cb3](https://github.com/nkapatos/mindweaver/commit/a4e9cb3cdce118c882297c740b0edb05dbd03861))
* **neoweaver:** add proto-&gt;TS-&gt;Lua type generation pipeline ([e36d675](https://github.com/nkapatos/mindweaver/commit/e36d67547f696c8681b4e706c7dcb4d55e582f20))
* **neoweaver:** add proto-to-lua type generation ([51c2a04](https://github.com/nkapatos/mindweaver/commit/51c2a044b54348583897b129f1dc6095dc9e91da))
* **neoweaver:** implement working proto-&gt;TS-&gt;Lua type generation ([af16393](https://github.com/nkapatos/mindweaver/commit/af1639372f3751ec958d09f0b70061eb56076cda))
* **neoweaver:** scaffold neovim client plugin ([4728355](https://github.com/nkapatos/mindweaver/commit/472835547ebc80af0355d431d2f16c95613f1b21))
* **nvim:** add server config and empty-note toggle ([8acdd41](https://github.com/nkapatos/mindweaver/commit/8acdd41dc7630bb4cb94125e6750d94763ea43d2))
* **nvim:** use NewNote endpoint with auto-generated titles ([796db97](https://github.com/nkapatos/mindweaver/commit/796db97ec66ed00b479a2cb47f9886be74f86bbe))


### Bug Fixes

* use camelCase in ReplaceNoteRequest and handle optional fields ([6078bfa](https://github.com/nkapatos/mindweaver/commit/6078bfa3629cc6d094c083ae57d8e4ed50b79339))


### Documentation

* **neoweaver:** update README using component template ([da42950](https://github.com/nkapatos/mindweaver/commit/da4295094cca0bc152e1120a222839d72d52334b))


### Maintenance

* **neoweaver:** update gitignore and remove stale comments ([0360e23](https://github.com/nkapatos/mindweaver/commit/0360e235c1ad6c03ebd1302e56592e66ad92229a))


### CI/CD

* add panvimdoc workflow ([88b0ff3](https://github.com/nkapatos/mindweaver/commit/88b0ff377c4bbe57b78cc11b566066b290658eb5))

## Changelog
