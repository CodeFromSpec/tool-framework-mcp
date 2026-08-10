# Changelog

## [6.1.0](https://github.com/CodeFromSpec/tool-framework-mcp/compare/v6.0.0...v6.1.0) (2026-08-10)


### Features

* add 'type: verdict' support ([#44](https://github.com/CodeFromSpec/tool-framework-mcp/issues/44)) ([abcbf8b](https://github.com/CodeFromSpec/tool-framework-mcp/commit/abcbf8b8f5032e3fa9ddc0b74b12ad57b055578b))
* add custom frontmatter field and reject unknown fields ([#41](https://github.com/CodeFromSpec/tool-framework-mcp/issues/41)) ([79b14b2](https://github.com/CodeFromSpec/tool-framework-mcp/commit/79b14b2451cf1e38b74740800504bc4dae158aa1))
* add frontmatter `type` field and default output path ([#43](https://github.com/CodeFromSpec/tool-framework-mcp/issues/43)) ([9d1acc0](https://github.com/CodeFromSpec/tool-framework-mcp/commit/9d1acc0bfb0ec23fd2ff2b7bdc3fd9437d69f2dd))
* add glob expansion for imports and input fields ([#46](https://github.com/CodeFromSpec/tool-framework-mcp/issues/46)) ([1e86457](https://github.com/CodeFromSpec/tool-framework-mcp/commit/1e864570111c17dab501f3eeeb907cd439552202))
* add wait_on frontmatter field with blocking support ([#47](https://github.com/CodeFromSpec/tool-framework-mcp/issues/47)) ([f5e875f](https://github.com/CodeFromSpec/tool-framework-mcp/commit/f5e875ff4767c3eec2e7f72e01cdeeb81d13ccef))
* multiple inputs ([#39](https://github.com/CodeFromSpec/tool-framework-mcp/issues/39)) ([d863834](https://github.com/CodeFromSpec/tool-framework-mcp/commit/d86383493fa3370327faed9a8b3f07b00c23c03c))
* Rename write_file MCP tool to write_artifact ([#45](https://github.com/CodeFromSpec/tool-framework-mcp/issues/45)) ([11b8c15](https://github.com/CodeFromSpec/tool-framework-mcp/commit/11b8c151f0cdf20d9a825d02acf569d46f99486a))
* renamed frontmatter field "depends_on" to "imports"  ([#38](https://github.com/CodeFromSpec/tool-framework-mcp/issues/38)) ([07cebf4](https://github.com/CodeFromSpec/tool-framework-mcp/commit/07cebf42f6fd18f9e81d0c774e4d9cc130009ffe))
* separate imports into dedicated &lt;references&gt; section in chain XML ([#40](https://github.com/CodeFromSpec/tool-framework-mcp/issues/40)) ([51ea0cb](https://github.com/CodeFromSpec/tool-framework-mcp/commit/51ea0cb942d60fd6b8ffdfd90748c7a2429c1cfc))


### Bug Fixes

* imports, go mod updated to v6 ([#36](https://github.com/CodeFromSpec/tool-framework-mcp/issues/36)) ([9e79d4a](https://github.com/CodeFromSpec/tool-framework-mcp/commit/9e79d4a2730f88e421a4fef8b1a9c7da9dcc8345))
* move domain/ specs to external/ ([#42](https://github.com/CodeFromSpec/tool-framework-mcp/issues/42)) ([683e679](https://github.com/CodeFromSpec/tool-framework-mcp/commit/683e679c680af0f2ffa07de1295d180967fa29e4))

## [6.0.0](https://github.com/CodeFromSpec/tool-framework-mcp/compare/v5.1.0...v6.0.0) (2026-08-07)


### ⚠ BREAKING CHANGES

* version 6, documentation ([#35](https://github.com/CodeFromSpec/tool-framework-mcp/issues/35))

### Features

* confine generation subagents to their target node via opaque tokens ([#34](https://github.com/CodeFromSpec/tool-framework-mcp/issues/34)) ([fc0d1df](https://github.com/CodeFromSpec/tool-framework-mcp/commit/fc0d1dfab6801e444db1bf5f4d4f7aca6770be2d))
* version 6, documentation ([#35](https://github.com/CodeFromSpec/tool-framework-mcp/issues/35)) ([a55699a](https://github.com/CodeFromSpec/tool-framework-mcp/commit/a55699aa0d7983cf06e2997a3fa2a9c7289914f5))


### Bug Fixes

* previous_instructions/previous_input double-nesting in load_chain XML ([#32](https://github.com/CodeFromSpec/tool-framework-mcp/issues/32)) ([29749e0](https://github.com/CodeFromSpec/tool-framework-mcp/commit/29749e091935f45027276b4a9bf0bd7a0804586c))

## [5.1.0](https://github.com/CodeFromSpec/tool-framework-mcp/compare/v5.0.0...v5.1.0) (2026-07-20)


### Features

* per-node dump files and prune_orphans tool ([#30](https://github.com/CodeFromSpec/tool-framework-mcp/issues/30)) ([5277d86](https://github.com/CodeFromSpec/tool-framework-mcp/commit/5277d865ca4054bc64cfa945521c1ec425588e7c))

## [5.0.0](https://github.com/CodeFromSpec/tool-framework-mcp/compare/v4.2.0...v5.0.0) (2026-07-01)


### ⚠ BREAKING CHANGES

* release v5, finally ([#28](https://github.com/CodeFromSpec/tool-framework-mcp/issues/28))

### Features

* release v5, finally ([#28](https://github.com/CodeFromSpec/tool-framework-mcp/issues/28)) ([676c136](https://github.com/CodeFromSpec/tool-framework-mcp/commit/676c1363a18e0a547d15f080165796dd634b383b))

## [4.2.0](https://github.com/CodeFromSpec/tool-framework-mcp/compare/v4.1.1...v4.2.0) (2026-06-30)


### Features

* add spec chain cache with disposition tracking ([#27](https://github.com/CodeFromSpec/tool-framework-mcp/issues/27)) ([f2ed342](https://github.com/CodeFromSpec/tool-framework-mcp/commit/f2ed342971d001e095b5aa1f9c93f64f802480ea))
* migrating repo to use code-from-spec v5 (from v4) ([#25](https://github.com/CodeFromSpec/tool-framework-mcp/issues/25)) ([cae1a9f](https://github.com/CodeFromSpec/tool-framework-mcp/commit/cae1a9f37453a83d3238cb33972d0ef41ee71e98))

## [4.1.1](https://github.com/CodeFromSpec/tool-framework-mcp/compare/v4.1.0...v4.1.1) (2026-06-30)


### Bug Fixes

* Remove path parameter from write_file tool ([#23](https://github.com/CodeFromSpec/tool-framework-mcp/issues/23)) ([0056377](https://github.com/CodeFromSpec/tool-framework-mcp/commit/0056377ca150ecece09908e90db2397c5537ce26))

## [4.1.0](https://github.com/CodeFromSpec/tool-framework-mcp/compare/v4.0.0...v4.1.0) (2026-06-30)


### Features

* tool supports v5, no cache support yet ([#21](https://github.com/CodeFromSpec/tool-framework-mcp/issues/21)) ([1b4d11a](https://github.com/CodeFromSpec/tool-framework-mcp/commit/1b4d11a1b967b24805276c81d1f7c1a7f2a329e9))

## [4.0.0](https://github.com/CodeFromSpec/tool-framework-mcp/compare/v3.0.0...v4.0.0) (2026-06-15)


### ⚠ BREAKING CHANGES

* artificially increasing major version to match the code_from_spec major version ([#19](https://github.com/CodeFromSpec/tool-framework-mcp/issues/19))

### Features

* artificially increasing major version to match the code_from_spec major version ([#19](https://github.com/CodeFromSpec/tool-framework-mcp/issues/19)) ([1f5b107](https://github.com/CodeFromSpec/tool-framework-mcp/commit/1f5b107e6a78f551d27bb5bbc672c2dca724aad5))

## [3.0.0](https://github.com/CodeFromSpec/tool-framework-mcp/compare/v2.0.0...v3.0.0) (2026-06-15)


### ⚠ BREAKING CHANGES

* artificially increasing major version ([#17](https://github.com/CodeFromSpec/tool-framework-mcp/issues/17))

### Features

* artificially increasing major version ([#17](https://github.com/CodeFromSpec/tool-framework-mcp/issues/17)) ([e9462c0](https://github.com/CodeFromSpec/tool-framework-mcp/commit/e9462c04234b8ea52ae598c2b5ad0d12cd81afa4))

## [2.0.0](https://github.com/CodeFromSpec/tool-framework-mcp/compare/v1.2.0...v2.0.0) (2026-06-15)


### ⚠ BREAKING CHANGES

* increase major version, update readme ([#16](https://github.com/CodeFromSpec/tool-framework-mcp/issues/16))

### Features

* implement Code from Spec v4 behavior ([#15](https://github.com/CodeFromSpec/tool-framework-mcp/issues/15)) ([b0d2cf4](https://github.com/CodeFromSpec/tool-framework-mcp/commit/b0d2cf43ed2284cad0f420b4f93bdb455d78eac9))
* increase major version, update readme ([#16](https://github.com/CodeFromSpec/tool-framework-mcp/issues/16)) ([bfff9e6](https://github.com/CodeFromSpec/tool-framework-mcp/commit/bfff9e625ca0f5a6e54532f39ce55db52eb1a7e3))


### Bug Fixes

* update README.md ([#13](https://github.com/CodeFromSpec/tool-framework-mcp/issues/13)) ([9dfee49](https://github.com/CodeFromSpec/tool-framework-mcp/commit/9dfee49d80a6371e1ace1ea2635624ced2b3d3df))

## [1.2.0](https://github.com/CodeFromSpec/tool-framework-mcp/compare/v1.1.0...v1.2.0) (2026-06-04)


### Features

* enforce #Public subsection rule and strip artifact tag from chain ([#11](https://github.com/CodeFromSpec/tool-framework-mcp/issues/11)) ([a90690a](https://github.com/CodeFromSpec/tool-framework-mcp/commit/a90690aa99b011756ccf9497e65be663e04d2009))

## [1.1.0](https://github.com/CodeFromSpec/tool-framework-mcp/compare/v1.0.0...v1.1.0) (2026-06-03)


### Features

* load_chain single response ([#10](https://github.com/CodeFromSpec/tool-framework-mcp/issues/10)) ([b5f2ad7](https://github.com/CodeFromSpec/tool-framework-mcp/commit/b5f2ad7dd0e6c1d79612e1ab01a138534ac98cd1))
* remove fragments and hash_fragment tool ([#7](https://github.com/CodeFromSpec/tool-framework-mcp/issues/7)) ([cbd90d9](https://github.com/CodeFromSpec/tool-framework-mcp/commit/cbd90d9c8e47b7e496863002ae09fda807555221))
* single output, remove fragments, add chain_hash and version tools ([#8](https://github.com/CodeFromSpec/tool-framework-mcp/issues/8)) ([7b55dda](https://github.com/CodeFromSpec/tool-framework-mcp/commit/7b55dda99912b598e3ac9bc3afedd667531d39a9))
* spec hardening, artifact tag neutralization, and full regeneration ([#9](https://github.com/CodeFromSpec/tool-framework-mcp/issues/9)) ([aed04bd](https://github.com/CodeFromSpec/tool-framework-mcp/commit/aed04bd5fdeff9db4a64ff18336fcb6324431974))


### Bug Fixes

* release_please files ([#5](https://github.com/CodeFromSpec/tool-framework-mcp/issues/5)) ([7cb1906](https://github.com/CodeFromSpec/tool-framework-mcp/commit/7cb19060faec2517031626b4e799c6b405b76969))

## 1.0.0 (2026-06-01)


### Features

* init ([#1](https://github.com/CodeFromSpec/tool-framework-mcp/issues/1)) ([c3e63d6](https://github.com/CodeFromSpec/tool-framework-mcp/commit/c3e63d6d3f94137372f13f1663a25032c67a8b17))
* updated version ([#3](https://github.com/CodeFromSpec/tool-framework-mcp/issues/3)) ([dcb9f5b](https://github.com/CodeFromSpec/tool-framework-mcp/commit/dcb9f5b5287c1c18a186a8358a76494da701a98c))


### Bug Fixes

* major version ([#4](https://github.com/CodeFromSpec/tool-framework-mcp/issues/4)) ([31ddb79](https://github.com/CodeFromSpec/tool-framework-mcp/commit/31ddb79933c4722c36fe0e3c10805844f08be8a3))
