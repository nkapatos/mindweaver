# Changelog

## [0.1.8](https://github.com/nkapatos/mindweaver/compare/v0.1.7...v0.1.8) (2026-01-07)


### Features

* **admin:** add first-run setup wizard ([39d7a4f](https://github.com/nkapatos/mindweaver/commit/39d7a4f27c609e873b47e3a05b4e504fc89f6247)), closes [#87](https://github.com/nkapatos/mindweaver/issues/87)
* **config:** consolidate database paths with Viper and DATA_DIR ([841fc3c](https://github.com/nkapatos/mindweaver/commit/841fc3c07581315adb5ff0a8cf9c8f75f6800e2b))
* **notes:** version body changes only, add relocated event ([aab96c4](https://github.com/nkapatos/mindweaver/commit/aab96c4836ddbcc6175230dbd8b3be24cbbc5707))
* **tags:** add ListNotesForTag and FindTags V3 endpoints ([c49e6e3](https://github.com/nkapatos/mindweaver/commit/c49e6e35fc056d0d88983890199b62e544f635cc)), closes [#39](https://github.com/nkapatos/mindweaver/issues/39)


### Bug Fixes

* resolve linter errors in config package ([d5e838b](https://github.com/nkapatos/mindweaver/commit/d5e838bf99d60ad6f59a52885a24b78ce610cfdd))
* resolve shadow variable lint error in main.go ([16fd6ab](https://github.com/nkapatos/mindweaver/commit/16fd6ab68aa146f93bb5fd9e88e9b5d7c2698080))
* **tags:** handle count errors per errcheck linter ([8103b6a](https://github.com/nkapatos/mindweaver/commit/8103b6a3f111e4b329cad3ba13a78db44d7d6f3b))

## [0.1.7](https://github.com/nkapatos/mindweaver/compare/v0.1.6...v0.1.7) (2026-01-02)


### Features

* **events:** add session tracking and wire all services to EventHub ([842cfb7](https://github.com/nkapatos/mindweaver/commit/842cfb7822b2f6471833f7d9c2fe91893cf572de))
* **events:** add SSE endpoint with EventHub for real-time updates ([9307ca7](https://github.com/nkapatos/mindweaver/commit/9307ca7a98565cf159f06011f06d9c946c1126ad))
* **events:** wire EventHub into NotesService for SSE notifications ([5367c50](https://github.com/nkapatos/mindweaver/commit/5367c507247256a1430c05d4492cc8bdfa573d4d))
* **notes:** add UpdateNote RPC for partial updates (PATCH) ([9b95fce](https://github.com/nkapatos/mindweaver/commit/9b95fce2be9f52285d7562c431ec988e63892dad))


### Bug Fixes

* update goreleaser config to remove deprecation warnings ([bfe88f7](https://github.com/nkapatos/mindweaver/commit/bfe88f71da2bfe7f6a2b4caa98e2523609163e35))


### Maintenance

* consolidate github config files under .github directory ([b3cbabb](https://github.com/nkapatos/mindweaver/commit/b3cbabb2fed04c3630314737c3d0d123bbdb53e3))
* remove unused REST search implementation ([e3ab281](https://github.com/nkapatos/mindweaver/commit/e3ab281b80a50d4ad6eb786992e20c1090e33a31))


### CI/CD

* add path filters to skip CI for non-code changes ([8c4d2ed](https://github.com/nkapatos/mindweaver/commit/8c4d2ed7776dda9d561e3bcaada1b6762ee6642c))

## [0.1.6](https://github.com/nkapatos/mindweaver/compare/v0.1.5...v0.1.6) (2025-12-27)


### Features

* add Docker Hub publishing to release workflow ([8de9e0c](https://github.com/nkapatos/mindweaver/commit/8de9e0cda9e09e235c0be7fa248a64b25630adde))


### Bug Fixes

* apply lint fixes for package comments and error formatting ([da4c56b](https://github.com/nkapatos/mindweaver/commit/da4c56b29c25cc2ab33c80dac56a3e62357e27df))
* correct errcheck exclusion syntax and add remaining exclusions ([174e127](https://github.com/nkapatos/mindweaver/commit/174e1276a84c6f8d2e2db5f2e282bb40c1a75c38))
* include internal/mind/gen in CI artifacts ([267385d](https://github.com/nkapatos/mindweaver/commit/267385dd55dd9d9b5d168f85132e041b173ab4b8))
* resolve golangci-lint warnings across codebase ([9c11a7b](https://github.com/nkapatos/mindweaver/commit/9c11a7b81fff84323b1b9b572fac11933deb4989))
* update air config to build from root instead of cmd/ ([8919c2a](https://github.com/nkapatos/mindweaver/commit/8919c2a12987cd6a2d09ce415327414ab53d6b1d))


### Maintenance

* exclude idiomatic error ignores from errcheck ([2d04a19](https://github.com/nkapatos/mindweaver/commit/2d04a19ad73662e94136d9c820fcd83132cef9ae))
* fix sql file broken from the sql formatter ([c807bf7](https://github.com/nkapatos/mindweaver/commit/c807bf7b30768bb1d072ab4b9f57bb8107a26639))
* lint fixes ([0c72734](https://github.com/nkapatos/mindweaver/commit/0c72734be7589dcf63d40062e308b416f7145d1f))
* migrate golangci-lint config to v2 and apply autofixes ([6c47a19](https://github.com/nkapatos/mindweaver/commit/6c47a199b1ea7c4f9691b127b6f5679b5f321cf1))
* **sql:** trim redundant comments and standardize section headers in mind store SQL files; move future work notes to GH issues ([ea2ff11](https://github.com/nkapatos/mindweaver/commit/ea2ff1194f6da1c8782606bf0808d93a3e2fc505))
* update mise tooling to add correct version of golancgci-lint for the project ([fe6e01d](https://github.com/nkapatos/mindweaver/commit/fe6e01d99d96b83e7707fcc32855ae05c2e08485))


### Refactoring

* consolidate Connect-RPC route registration into bootstrap ([5a82ff4](https://github.com/nkapatos/mindweaver/commit/5a82ff452dded3709d419f7368a6ca66628ef405))
* extract validation interceptor to shared package ([e9b347b](https://github.com/nkapatos/mindweaver/commit/e9b347bb70fc97b87a6519dbc7b6e3d0351da3be))
* standardize error handling in handlers and services ([516da2c](https://github.com/nkapatos/mindweaver/commit/516da2c0a2fc272bf75a03b4962e59dc9426bb52))


### CI/CD

* optimize workflow structure with build and lint jobs ([a7ab0d1](https://github.com/nkapatos/mindweaver/commit/a7ab0d104eb5d25541a60a9fe238807aa0999b3b))
* update golangci lint version to match workspace go version reqs ([61248d1](https://github.com/nkapatos/mindweaver/commit/61248d158bc819ad6fcaea20b6f9a7e3463b7285))

## [0.1.5](https://github.com/nkapatos/mindweaver/compare/v0.1.4...v0.1.5) (2025-12-26)


### Bug Fixes

* configure release-please for standalone repo tags ([7b45b54](https://github.com/nkapatos/mindweaver/commit/7b45b54e842b99d44ef20b70de119e9cae0e180c))


### Maintenance

* clean up buf.gen.yaml for standalone repo ([b996025](https://github.com/nkapatos/mindweaver/commit/b996025d9e8dd045d0f50bdfb33f1e2089c62cf1))
* clean up config for standalone repo ([465ef46](https://github.com/nkapatos/mindweaver/commit/465ef466fae20e0759a5322bde1512f2e9c777d2))
* comments cleanup ([361bfda](https://github.com/nkapatos/mindweaver/commit/361bfdafd568f3aa5a2a4cfeef3c7bc48fae856d))
* remove neoweaver sync workflow ([fc2d74d](https://github.com/nkapatos/mindweaver/commit/fc2d74d8026e57fde90b6209a5a067d847a204f1))
* update issue templates for standalone server repo ([4fe0954](https://github.com/nkapatos/mindweaver/commit/4fe095463daf61b30a0b9e9733f7b9afc6f2a895))
* update task paths for standalone repo ([90a5bff](https://github.com/nkapatos/mindweaver/commit/90a5bfffb35cc710101835b1f54737107a328b8d))


### Refactoring

* update import paths for standalone repo ([924d746](https://github.com/nkapatos/mindweaver/commit/924d746536c869a7eb73766d1153c02446e45a48))


### CI/CD

* clean up workflows for standalone repo ([d37ce1f](https://github.com/nkapatos/mindweaver/commit/d37ce1f37ca8827f264a87dbc164036c894824eb))
* simplify release-please workflow for standalone repo ([711ae81](https://github.com/nkapatos/mindweaver/commit/711ae818f04818c74a387acc83117f3b5d22d6df))

## [0.1.4](https://github.com/nkapatos/mindweaver/compare/mindweaver/v0.1.3...mindweaver/v0.1.4) (2025-12-25)


### CI/CD

* disable CGO and add code generation step before GoReleaser ([1db03a8](https://github.com/nkapatos/mindweaver/commit/1db03a8226a0353bf8a147d8bf0bde6ec0a43373))

## [0.1.3](https://github.com/nkapatos/mindweaver/compare/mindweaver/v0.1.2...mindweaver/v0.1.3) (2025-12-25)


### Features

* **mindweaver:** implement notes:find endpoint for global search ([ebf635b](https://github.com/nkapatos/mindweaver/commit/ebf635b410c385fa4f456a2056bbbec2b3500f61))
* **neoweaver:** integrate notes:find API for search picker ([ebf635b](https://github.com/nkapatos/mindweaver/commit/ebf635b410c385fa4f456a2056bbbec2b3500f61))


### CI/CD

* **mindweaver:** optimize CI workflow to reduce buf rate limits ([ea0d16b](https://github.com/nkapatos/mindweaver/commit/ea0d16b8bd79c88c0590638b7ed6aaae6a0af0bd))

## [0.1.2](https://github.com/nkapatos/mindweaver/compare/mindweaver/v0.1.1...mindweaver/v0.1.2) (2025-12-23)


### Features

* **mindweaver:** add error details for etagMissmatch for connet error response ([1439c1c](https://github.com/nkapatos/mindweaver/commit/1439c1ca0d52dadd8b0a6c13642cc690e5d6809a))
* **mindweaver:** implement field masking in ListNotes handler ([666e079](https://github.com/nkapatos/mindweaver/commit/666e079f9a0af5b856d299e53200ed53d7d6ec85))


### Maintenance

* **mindweaver:** remove v1 api types from note types pkg ([24aa562](https://github.com/nkapatos/mindweaver/commit/24aa562aeb173fd8aac7f8a74015327d78c9dd42))
* **mindweaver:** setup GitHub workflow and clean TODO comments ([b2d88ed](https://github.com/nkapatos/mindweaver/commit/b2d88ed28c2fd6044fc003d8ace202e28c1e8d5b))
* tooling in mise and ignore gen, tmp dirs from git ([fe994c4](https://github.com/nkapatos/mindweaver/commit/fe994c4ff92ab6a39ff5c2625f4d63fc13a67c87))


### Refactoring

* **errors:** remove IsNotFoundError helper, use sql.ErrNoRows directly ([049e3fa](https://github.com/nkapatos/mindweaver/commit/049e3fa5d6f6766f2749d62bf82c5e24526a9cc4))
* **mind:** remove _v3 suffix from all packages except search ([93c6540](https://github.com/nkapatos/mindweaver/commit/93c6540d97afabfcbed0a80c1ba189e598cefd5e))
* **mind:** remove obsolete shared/routes package ([fdb3dbe](https://github.com/nkapatos/mindweaver/commit/fdb3dbe7f046a1ecae4676e7634b6049223dd2d6))
* **mind:** standardize error handling across all handlers ([f803674](https://github.com/nkapatos/mindweaver/commit/f803674692c01f58c765ac483ee2af1c14625300))
* **mindweaver:** consolidate error handling in shared/errors package ([97da7a8](https://github.com/nkapatos/mindweaver/commit/97da7a8ad331067a37acdac5f82d7d92614d74dc))
* **mindweaver:** replace manual sql.Null* constructions with centralized utils helpers ([48018ed](https://github.com/nkapatos/mindweaver/commit/48018ed00ae7a0b83563bfcb9951fcc7b040e891))


### Tests

* **mindweaver:** add Bruno test for field masking in ListNotes ([bc71f25](https://github.com/nkapatos/mindweaver/commit/bc71f25dd8512575fdca4026ff51a7b2c1f9b55a))


### CI/CD

* component-scoped filters, Go cache; neoweaver release sync ([#20](https://github.com/nkapatos/mindweaver/issues/20)) ([90cfa9e](https://github.com/nkapatos/mindweaver/commit/90cfa9edbd70bf78ab8227a84baf2a2eb9b71d96))

## [0.1.1](https://github.com/nkapatos/mindweaver/compare/mindweaver/v0.1.0...mindweaver/v0.1.1) (2025-12-16)


### Features

* establish component-specific structure ([1af0de2](https://github.com/nkapatos/mindweaver/commit/1af0de278fa3feef5bafd9c11b50081b56e001e1))
* **mindweaver:** add NewNote endpoint for auto-generated note creation ([443e4c7](https://github.com/nkapatos/mindweaver/commit/443e4c75f0e64aae73de431e24915082289b2a3d))


### Bug Fixes

* correct Go module paths for proto generation ([63816ea](https://github.com/nkapatos/mindweaver/commit/63816ea3caf0a33b5c3eadceb03470afffd62285))
* **mindweaver:** add store generation dependency to build task ([c8f5269](https://github.com/nkapatos/mindweaver/commit/c8f5269569be36db1425105726cdcd61013598e5))
* **mindweaver:** correct proto generation paths and imports ([80c86a4](https://github.com/nkapatos/mindweaver/commit/80c86a41a0b6987acf254a72626b404f08322315))
* **mindweaver:** remove packages/pkg from workspace setup ([4a09804](https://github.com/nkapatos/mindweaver/commit/4a098047fccb6f3a93cfcbfd409dc7df3a1dc4b0))
* **mindweaver:** remove workspace:setup dependency from build task ([6c328e5](https://github.com/nkapatos/mindweaver/commit/6c328e59d8eb98c562859a019fdcbf10c8ead3cf))


### Documentation

* **mindweaver:** update README using component template ([71e8bc7](https://github.com/nkapatos/mindweaver/commit/71e8bc79c650296a3637e821ab5be7173ef98e2f))
* move tool prerequisites to appropriate levels ([d9c7712](https://github.com/nkapatos/mindweaver/commit/d9c771293d3bcf5f18b7f9ef81ca456f4c2f5f41))
* update prerequisites and add component-level documentation ([9b69622](https://github.com/nkapatos/mindweaver/commit/9b69622199af3af5266641baf2f38a20a8469701))


### Maintenance

* create monorepo directory structure ([841b2b0](https://github.com/nkapatos/mindweaver/commit/841b2b078a7c57b2baf491c067c5ccb964f78e17))
* move config files to monorepo structure ([10888fc](https://github.com/nkapatos/mindweaver/commit/10888fc19edfbd98d5a261ca5b4e6c4f75b53da9))
* regenerate code and fix module references for monorepo ([2661a9c](https://github.com/nkapatos/mindweaver/commit/2661a9c67ffff3a523e92d9d1bff5c5a58dbf1ae))
* remove root go.mod and use Task runner in CI ([b665e5b](https://github.com/nkapatos/mindweaver/commit/b665e5bdc2badc213bedd945060405993dd36377))
* reset version to 0.0.0 for fresh monorepo start ([617d347](https://github.com/nkapatos/mindweaver/commit/617d347a787384723fab4acc6f26075be364762a))
* setup go workspace with separate modules ([60446c9](https://github.com/nkapatos/mindweaver/commit/60446c9a472467e14595ad3b9f1bbb35a4ff60fe))
* update release and CI workflows for monorepo structure ([61f4024](https://github.com/nkapatos/mindweaver/commit/61f40247ca781d918e5632b260a2be821017bbf1))
* update task variable defaults for monorepo paths ([5d5fd80](https://github.com/nkapatos/mindweaver/commit/5d5fd8075221fb41f988afd5c1a4eb040a20844e))


### Refactoring

* **mindweaver:** add db:init task and improve task structure ([4340c3b](https://github.com/nkapatos/mindweaver/commit/4340c3be3a5c7cada1133568f573bc8913f8a517))
* **mindweaver:** add proto generation tasks and ignore generated files ([a869f26](https://github.com/nkapatos/mindweaver/commit/a869f26a7073f4e78974abbd615e6cebca439b38))
* **mindweaver:** move packages/pkg to packages/mindweaver/shared ([a16d34f](https://github.com/nkapatos/mindweaver/commit/a16d34f5912216c2e50d4d939c07ee665ba92d55))
* **mindweaver:** relocate proto generation to consumer directory ([b16c806](https://github.com/nkapatos/mindweaver/commit/b16c806e70770e72efb84237070edb63e84275cd))
* **mindweaver:** update import paths to use shared package location ([253e0ca](https://github.com/nkapatos/mindweaver/commit/253e0caec416967b40df866aa1904ea7b63214cf))
* move files to monorepo structure ([57446c7](https://github.com/nkapatos/mindweaver/commit/57446c7fc5a76229d04180509cc18022ebcf805b))
* update import paths for monorepo structure ([8fd25b8](https://github.com/nkapatos/mindweaver/commit/8fd25b80b79e4a4f02687d458fba8c24acf7b143))


### CI/CD

* automate Go workspace setup via Task runner ([459332a](https://github.com/nkapatos/mindweaver/commit/459332abd5cf5c1d9882558aa5e8d09f402d5815))
* **mindweaver:** add path filtering and conditional jobs to CI workflow ([8cd4c37](https://github.com/nkapatos/mindweaver/commit/8cd4c375138d290fea00a71c9c7340fd74a36768))


### Build System

* **mindweaver:** remove unnecessary replace directive ([b213c46](https://github.com/nkapatos/mindweaver/commit/b213c46183fa5609706d5cbea480036c71a56548))
* update task files for mindweaver package to have the correct dir context ([1c52b49](https://github.com/nkapatos/mindweaver/commit/1c52b49f3b8c77e5cc6fc8c34d730d12ec7f1968))
* update task paths for packages/mindweaver context ([51b67dc](https://github.com/nkapatos/mindweaver/commit/51b67dcb494b4520f40332fd9bfc7afa303a0c0d))

## Changelog
