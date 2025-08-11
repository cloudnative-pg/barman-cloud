# Changelog

## [0.3.3](https://github.com/cloudnative-pg/barman-cloud/compare/v0.3.2...v0.3.3) (2025-08-11)


### Bug Fixes

* **deps:** update module github.com/cloudnative-pg/machinery to v0.3.1 ([#142](https://github.com/cloudnative-pg/barman-cloud/issues/142)) ([276b8dd](https://github.com/cloudnative-pg/barman-cloud/commit/276b8dd9cb9c99062bba20399ccaf996e16e4c86))

## [0.3.2](https://github.com/cloudnative-pg/barman-cloud/compare/v0.3.1...v0.3.2) (2025-08-11)


### Bug Fixes

* **deps:** update all non-major go dependencies ([#109](https://github.com/cloudnative-pg/barman-cloud/issues/109)) ([6bb2b14](https://github.com/cloudnative-pg/barman-cloud/commit/6bb2b14d3e54d9d9ec27377430f93f7c7077d34e))
* **deps:** update all non-major go dependencies ([#136](https://github.com/cloudnative-pg/barman-cloud/issues/136)) ([ee0a802](https://github.com/cloudnative-pg/barman-cloud/commit/ee0a802628a1bdc9232837fa5853b0c14700b6fb))
* **deps:** update kubernetes packages to v0.33.1 ([#117](https://github.com/cloudnative-pg/barman-cloud/issues/117)) ([da740ad](https://github.com/cloudnative-pg/barman-cloud/commit/da740adb71f97e65c9471276c20bcf5b3972eeba))
* **deps:** update kubernetes packages to v0.33.3 ([#132](https://github.com/cloudnative-pg/barman-cloud/issues/132)) ([bcd06b7](https://github.com/cloudnative-pg/barman-cloud/commit/bcd06b750a126850852b4d3a636da0de5e1af1b0))
* **deps:** update module sigs.k8s.io/controller-runtime to v0.21.0 ([#125](https://github.com/cloudnative-pg/barman-cloud/issues/125)) ([96af481](https://github.com/cloudnative-pg/barman-cloud/commit/96af4817196af0b7df154bbb2e480f9571c3ffb1))

## [0.3.1](https://github.com/cloudnative-pg/barman-cloud/compare/v0.3.0...v0.3.1) (2025-03-27)


### Bug Fixes

* **deps:** update all non-major go dependencies ([#92](https://github.com/cloudnative-pg/barman-cloud/issues/92)) ([1e425f4](https://github.com/cloudnative-pg/barman-cloud/commit/1e425f4272b7b36b41c94ea89bbbb863ee7ed864))
* **deps:** update kubernetes packages to v0.32.3 ([#80](https://github.com/cloudnative-pg/barman-cloud/issues/80)) ([aed9756](https://github.com/cloudnative-pg/barman-cloud/commit/aed9756f643f314fb7c6f6fd074dc7b44fd95872))
* **deps:** update module github.com/cloudnative-pg/machinery to v0.2.0 ([#101](https://github.com/cloudnative-pg/barman-cloud/issues/101)) ([a40877d](https://github.com/cloudnative-pg/barman-cloud/commit/a40877d28dcdd403f0657640693df5bde3e98c4b))
* **deps:** update module github.com/onsi/ginkgo/v2 to v2.23.1 ([#91](https://github.com/cloudnative-pg/barman-cloud/issues/91)) ([42588ca](https://github.com/cloudnative-pg/barman-cloud/commit/42588ca339ce6ce178f4ec1a1a301acb933785e2))
* **deps:** update module sigs.k8s.io/controller-runtime to v0.20.4 ([#77](https://github.com/cloudnative-pg/barman-cloud/issues/77)) ([981604e](https://github.com/cloudnative-pg/barman-cloud/commit/981604e2d24940f84f6f88088c920ce3e37eb172))
* remove lz4, xz, and zstd compression for backups ([#89](https://github.com/cloudnative-pg/barman-cloud/issues/89)) ([d53d05b](https://github.com/cloudnative-pg/barman-cloud/commit/d53d05b8f023dab6be20bd9bbb0b592470ef8662)), closes [#88](https://github.com/cloudnative-pg/barman-cloud/issues/88)
* use a fixed golangci-lint version ([#94](https://github.com/cloudnative-pg/barman-cloud/issues/94)) ([b1df782](https://github.com/cloudnative-pg/barman-cloud/commit/b1df7824d821742a26cd03651ed2ab6a1426e397))

## [0.3.0](https://github.com/cloudnative-pg/barman-cloud/compare/v0.2.0...v0.3.0) (2025-03-18)


### Features

* add lz4, xz, and zstd compression ([#82](https://github.com/cloudnative-pg/barman-cloud/issues/82)) ([6848fd4](https://github.com/cloudnative-pg/barman-cloud/commit/6848fd45696b2eb66ea2b40b4c3a006e64028bcc))

## [0.2.0](https://github.com/cloudnative-pg/barman-cloud/compare/v0.1.0...v0.2.0) (2025-03-13)


### Features

* allow using a custom directory for CA certificates ([#78](https://github.com/cloudnative-pg/barman-cloud/issues/78)) ([3fc2d78](https://github.com/cloudnative-pg/barman-cloud/commit/3fc2d78dca9ab469f7460f1faaa975b802baab95))


### Bug Fixes

* **deps:** update module github.com/onsi/ginkgo/v2 to v2.23.0 ([#76](https://github.com/cloudnative-pg/barman-cloud/issues/76)) ([72ba30c](https://github.com/cloudnative-pg/barman-cloud/commit/72ba30c8e72d8c71aeae594f72ccd5ce6b2b6421))

## 0.1.0 (2025-02-26)


### Features

* ability to defaultAzureCredential for azure-blob-storage  ([#64](https://github.com/cloudnative-pg/barman-cloud/issues/64)) ([1a6b98d](https://github.com/cloudnative-pg/barman-cloud/commit/1a6b98ded711a39c01042402d04b2cba7e48932d)), closes [#59](https://github.com/cloudnative-pg/barman-cloud/issues/59)
* add webhook validator `ValidateBackupConfiguration` ([#14](https://github.com/cloudnative-pg/barman-cloud/issues/14)) ([7b60289](https://github.com/cloudnative-pg/barman-cloud/commit/7b60289361469ddf5ef1167b91958cab4394e3e3))
* initial import ([#2](https://github.com/cloudnative-pg/barman-cloud/issues/2)) ([44955af](https://github.com/cloudnative-pg/barman-cloud/commit/44955af09635c3dc0fffaa005d5a6274540bf405))
* make barman catalog compatible with the common backup interface ([#16](https://github.com/cloudnative-pg/barman-cloud/issues/16)) ([7b615ee](https://github.com/cloudnative-pg/barman-cloud/commit/7b615eefebac00b2b2b6d6edf7631485d7c6c8d3))
* support ISO format for dates in the barman-cloud output ([#49](https://github.com/cloudnative-pg/barman-cloud/issues/49)) ([d99f49b](https://github.com/cloudnative-pg/barman-cloud/commit/d99f49ba79d7059fa16ad54ff34fdda5d2286ced))


### Bug Fixes

* **deps:** update all non-major go dependencies ([#41](https://github.com/cloudnative-pg/barman-cloud/issues/41)) ([ae6c240](https://github.com/cloudnative-pg/barman-cloud/commit/ae6c2408bd14ebdc8443322988f3a5ab7e9e4730))
* **deps:** update all non-major go dependencies ([#43](https://github.com/cloudnative-pg/barman-cloud/issues/43)) ([10ef19b](https://github.com/cloudnative-pg/barman-cloud/commit/10ef19b66efec518beaf55977dece9680b45f95d))
* **deps:** update github.com/cloudnative-pg/machinery digest to 01cb70a ([#15](https://github.com/cloudnative-pg/barman-cloud/issues/15)) ([4e3e45c](https://github.com/cloudnative-pg/barman-cloud/commit/4e3e45cb0a5b1504c6efc9c2d7c3322b11ff35ba))
* **deps:** update github.com/cloudnative-pg/machinery digest to 6c50ae1 ([#10](https://github.com/cloudnative-pg/barman-cloud/issues/10)) ([70ddc94](https://github.com/cloudnative-pg/barman-cloud/commit/70ddc94656cce689c0766a2225d73aff388f1b53))
* **deps:** update github.com/cloudnative-pg/machinery digest to 9dd62b9 ([#21](https://github.com/cloudnative-pg/barman-cloud/issues/21)) ([bca019e](https://github.com/cloudnative-pg/barman-cloud/commit/bca019ea378221a45d587617063fe05cecd37ca5))
* **deps:** update github.com/cloudnative-pg/machinery digest to c27747f ([#27](https://github.com/cloudnative-pg/barman-cloud/issues/27)) ([71ee406](https://github.com/cloudnative-pg/barman-cloud/commit/71ee4065f9c76904490a31b28b8f598982f10e39))
* **deps:** update kubernetes packages to v0.32.2 ([#12](https://github.com/cloudnative-pg/barman-cloud/issues/12)) ([cfcb8af](https://github.com/cloudnative-pg/barman-cloud/commit/cfcb8af064e78f7b21ac11a3be6d7871a9610d0e))
* **deps:** update module github.com/cloudnative-pg/machinery to v0.1.0 ([#70](https://github.com/cloudnative-pg/barman-cloud/issues/70)) ([cb9c4f4](https://github.com/cloudnative-pg/barman-cloud/commit/cb9c4f4985476e4658fa5c814cfdc28ef276acb3))
* **deps:** update module sigs.k8s.io/controller-runtime to v0.20.2 ([#13](https://github.com/cloudnative-pg/barman-cloud/issues/13)) ([10d088c](https://github.com/cloudnative-pg/barman-cloud/commit/10d088c910ea5da92a39b1021790239b8890dad2))
* notify in the logs about backup completion ([#34](https://github.com/cloudnative-pg/barman-cloud/issues/34)) ([44f56f7](https://github.com/cloudnative-pg/barman-cloud/commit/44f56f711a5caa4f03ee5a971c0c7c75267ae632))
* **PITR:** compare TargetLSN with backup EndLSN instead of BeginLSN ([#56](https://github.com/cloudnative-pg/barman-cloud/issues/56)) ([018944b](https://github.com/cloudnative-pg/barman-cloud/commit/018944b15fd48aa8ae7dffa86829d49d1788ad9f)), closes [#6536](https://github.com/cloudnative-pg/barman-cloud/issues/6536)
* use RFC3339 format to parse ISO times ([#55](https://github.com/cloudnative-pg/barman-cloud/issues/55)) ([134c7de](https://github.com/cloudnative-pg/barman-cloud/commit/134c7de4954a53407d9da8ac3018ca689144bc41))
