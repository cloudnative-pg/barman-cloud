# Changelog

## 0.1.0 (2025-01-16)


### Features

* add webhook validator `ValidateBackupConfiguration` ([#14](https://github.com/cloudnative-pg/barman-cloud/issues/14)) ([7b60289](https://github.com/cloudnative-pg/barman-cloud/commit/7b60289361469ddf5ef1167b91958cab4394e3e3))
* initial import ([#2](https://github.com/cloudnative-pg/barman-cloud/issues/2)) ([44955af](https://github.com/cloudnative-pg/barman-cloud/commit/44955af09635c3dc0fffaa005d5a6274540bf405))
* make barman catalog compatible with the common backup interface ([#16](https://github.com/cloudnative-pg/barman-cloud/issues/16)) ([7b615ee](https://github.com/cloudnative-pg/barman-cloud/commit/7b615eefebac00b2b2b6d6edf7631485d7c6c8d3))
* support ISO format for dates in the barman-cloud output ([#49](https://github.com/cloudnative-pg/barman-cloud/issues/49)) ([d99f49b](https://github.com/cloudnative-pg/barman-cloud/commit/d99f49ba79d7059fa16ad54ff34fdda5d2286ced))


### Bug Fixes

* **deps:** update all non-major go dependencies ([#41](https://github.com/cloudnative-pg/barman-cloud/issues/41)) ([ae6c240](https://github.com/cloudnative-pg/barman-cloud/commit/ae6c2408bd14ebdc8443322988f3a5ab7e9e4730))
* **deps:** update github.com/cloudnative-pg/machinery digest to 01cb70a ([#15](https://github.com/cloudnative-pg/barman-cloud/issues/15)) ([4e3e45c](https://github.com/cloudnative-pg/barman-cloud/commit/4e3e45cb0a5b1504c6efc9c2d7c3322b11ff35ba))
* **deps:** update github.com/cloudnative-pg/machinery digest to 6c50ae1 ([#10](https://github.com/cloudnative-pg/barman-cloud/issues/10)) ([70ddc94](https://github.com/cloudnative-pg/barman-cloud/commit/70ddc94656cce689c0766a2225d73aff388f1b53))
* **deps:** update github.com/cloudnative-pg/machinery digest to 9dd62b9 ([#21](https://github.com/cloudnative-pg/barman-cloud/issues/21)) ([bca019e](https://github.com/cloudnative-pg/barman-cloud/commit/bca019ea378221a45d587617063fe05cecd37ca5))
* **deps:** update github.com/cloudnative-pg/machinery digest to c27747f ([#27](https://github.com/cloudnative-pg/barman-cloud/issues/27)) ([71ee406](https://github.com/cloudnative-pg/barman-cloud/commit/71ee4065f9c76904490a31b28b8f598982f10e39))
* notify in the logs about backup completion ([#34](https://github.com/cloudnative-pg/barman-cloud/issues/34)) ([44f56f7](https://github.com/cloudnative-pg/barman-cloud/commit/44f56f711a5caa4f03ee5a971c0c7c75267ae632))
* **PITR:** compare TargetLSN with backup EndLSN instead of BeginLSN ([#56](https://github.com/cloudnative-pg/barman-cloud/issues/56)) ([018944b](https://github.com/cloudnative-pg/barman-cloud/commit/018944b15fd48aa8ae7dffa86829d49d1788ad9f)), closes [#6536](https://github.com/cloudnative-pg/barman-cloud/issues/6536)
* use RFC3339 format to parse ISO times ([#55](https://github.com/cloudnative-pg/barman-cloud/issues/55)) ([134c7de](https://github.com/cloudnative-pg/barman-cloud/commit/134c7de4954a53407d9da8ac3018ca689144bc41))
