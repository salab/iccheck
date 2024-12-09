<!-- Keep a Changelog guide -> https://keepachangelog.com -->

# ICCheck ChangeLog

## [Unreleased]

## [0.8.0] - 2024-12-09

- 1d271d8e70cc55c86299e268623bb59f9d06fb14 v0.8.0
- 71161970c104bfa6866e900cc179d61aadad806f docs: update readme
- a19d7255e80a83dec1fe912873b4cc4e0e18931e feat: raw "search" command to search for clones
- 66d6d81322426b9beaf28a09cba73e5046ca2e43 refactor: split query extraction function
- ffa14c0acfd95f34416c3b734c300c39f9b7528c Merge pull request #19 from salab/intellij-changelog-update-0.7.6
- 4f4341a02051baf0f37eb22339dd3cc9cbacba86 IntelliJ Changelog Update - 0.7.6

## [0.7.6] - 2024-12-05

- 7009d2d07063556f2f87559834deca76473f2205 v0.7.6
- 99c8d582f8da763e08ff22cbbcd7aa32290db14c fix: perf: Slice indexing
- ec05b76c67b2eebf421160b27fd4d588cc19d712 Merge pull request #18 from salab/intellij-changelog-update-0.7.5
- f3e64ba2d839f991cd063bcad37e836a68b93b9a IntelliJ Changelog Update - 0.7.5

## [0.7.5] - 2024-12-05

- 0f9c2805132d6737b874e63f4faf948b27f198d8 v0.7.5
- d9dd9eac2c16f7b63b23d8a627c0d41ca6bed2ec fixup! fix: Skip region included context lines (+-4 lines)
- 99308260977081c990e9464fee0fea76680d420a fix: Skip region included context lines (+-4 lines)
- 19edddd7f122708c55455908c71f6661d37c47c4 Merge pull request #17 from salab/intellij-changelog-update-0.7.4
- 0379fb4989b44f8146191e0c06fdbad3fb3aba30 IntelliJ Changelog Update - 0.7.4

## [0.7.4] - 2024-12-05

- 9da9b61db634a6eda5d98e555189c54a074983cf v0.7.4
- da9a64f7d24f90b47c7d80619af192fbdd5b831c perf: Skip go bounds-check (+25% perf)
- 7207ea54b5bcd33aa1d7b292112a24aa064cd76c perf: Branchless code (+15% perf)
- 4ee7246fc50008ddb731b93520cbffb7ebc70ec4 Merge pull request #16 from salab/intellij-changelog-update-0.7.3
- 7ff173579627229595d8323640773eb51077b947 IntelliJ Changelog Update - 0.7.3

## [0.7.3] - 2024-12-05

- 3a27562b22bb17508514b9048698118d7432b45c v0.7.3
- 7afd704505b799af7aea810daa2103c25b4619de perf: Use forked go-git and disable workaround
- 68e6eb966ce709e76df0bf4925f059f78c0345d0 chores: test code: runtime/pprof
- a8ef551b6ee86af5a1ea383cc76b6e285b2b0442 chores: Increase default timeout
- e174382975b8a6ea1d2243a0c280140f59266e43 Merge pull request #15 from salab/intellij-changelog-update-0.7.2
- 61ff2c3e05698eeb0f358294abe34ba235abdb25 IntelliJ Changelog Update - 0.7.2

## [0.7.2] - 2024-12-02

- b846b32a8d8b9653d8448263569098c1588777d5 v0.7.2
- 50c7c71832873a8ae993c26a26402a359c2c1866 perf(lsp): Cache IsBinary() calculation
- f59ad839626978b48ebffc2277770bd4ee5da1b0 chores: Fix doc
- fc7f13af2f159c5cd2accc27b65ec623733ee263 chores: Improve logging
- eb2c65fcab261eadae0e50eb840753103990783f fix: Renamed file detection was not working
- 31d9aaa976d3c13a643a999ebdabc88a7d3f90b9 chores: Add --version command / update README
- 7e68edc1d14be0ebc1bf1b305787453d539e73e3 fix: Use HEAD tree to correctly calculate status/diff
- 73bf191c015bf296d6d7e2b1bf243624a96de911 Merge pull request #14 from salab/intellij-changelog-update-0.7.1
- b2d1f3f11f6fd8ab2a9e8a6843d42bb4c74e41f0 IntelliJ Changelog Update - 0.7.1

## [0.7.1] - 2024-11-27

- 7bcdf4a70c22ece3d99632baa87d8063a447092b v0.7.1
- 47597bb9fdfff0baf06efdda82cd1eba6510963a fix: Load system gitignore patterns
- d6eb9a3271ac94dc9db6fbe4749f59dc92747619 fix: concurrent writes / refactor overlay
- e6243cc66188fd5980cf56e1b3c2362f955fae4f perf: Preload files in go-git commit object
- 30a87a9e3ae281e9bd7af6cb1f3bd1ef9939297a fix: fixup perf: use uint16
- 41d6d535fd4ede0dcaee499a90568d8f36e3cdf6 perf: Optimize bigram intersection
- a4fa46a0be76d54c4ccb9c33e73cf22b25b38b9f feat: Add algorithm flag and adapt ncdsearch algorithm
- a44c8aa2b89a92a2f3c73c2d243c7712fa78dc10 Merge pull request #13 from salab/intellij-changelog-update-0.7.0
- d1aa5177c3a0323773399a3a3ada7bc9b7f673fd IntelliJ Changelog Update - 0.7.0

## [0.7.0] - 2024-11-26

- 3f22f97aac4fbe48e4cf69a53f79c6f8f7f51a5c fix: go mod tidy
- 4e548c15a6e84ec2be10824eba2e1a75ea8874ff v0.7.0
- c0e619b7546332f5bbb81946e61d711ae6ac61ba impr: README, add visual to top
- 0ae0c8ca04405559f65698aa21c3da8e690350e2 chores: Update packages, README
- d46bf7e87fb62877f0052534659589ecc4c8aa41 impr: Display version when installed via go install
- b9b092ce1ebcc9dfa2b8746189849a24c739e1a3 chores: Use buildinfo to display metadata
- 757b539b344551f9ce863d10a5de5c0192dbdbd8 impr: Auto suppress info logs by default on pipe
- 001c1fe3e88df803517cdb6e6d23c4fd771b6670 impr: Concise printers
- c29a7ec25c173699596580b73ce66e79b55308b9 impr: More concise logging
- 244dbb704a7068ef3d98e6eea98e6ad3f9c7d6d2 fix: yaml syntax in readme
- 208586e10c94d10a0e9eae9e0103dd921e1df8b4 impr(log): Concise logging
- f692a4aaa3ee689423a9aa21ddd7e76c479c5464 impr: Auto detect toplevel git / support bare
- e6a858f7f21881b2af945179ab1349ede2a0990e impr: Smarter default refs detection
- 36ab76b5d3d88d9292c08231b931753e90e36e66 chores: Fix diff / crlf problem
- 2e8008e4b69d00d5cf4cfaea6a1df1182baf366d chores: Update go-git
- d5168e5d4688669d41e9b24e5f54fdb845f9c27b chores(lsp:intellij): Suppress codeAction error logs
- 9242d239545683be3cd3b476abeef5be968d1595 fix: Golang default ignore rule
- 2eedf919392706049cb379630fe78add01424154 feat: Allow ignoring files/contents by regexp
- 0bc3a56d3234bdc59c74344796660dce7cafec69 chores: Remove obsolete test code
- fcd14f9af9bf1bf6f0e9d5b39bca56ce63534022 fix: version/revision embedding
- 03a4506ccc208548aa49b8595fb0741fb1b105d7 fix vscode publish ci
- 44edd7fa24a27a353cb71ede8161d4ca51f0071d Merge pull request #12 from salab/intellij-changelog-update-0.6.0
- d73becff7053cd6968ee61f97efe55f6d62f082a IntelliJ Changelog Update - 0.6.0

## [0.6.0] - 2024-11-17

- b51d2ea6b2673f72b2f91b880916f865894853e1 v0.6.0
- 9a676da2d90afc0c479d7de74aafdec0c0b9bded feat(lsp/vscode): Prefer binary on PATH if installed
- 02d1f48ba44a2fe80a3bd2b99272d5215ee8a9b4 feat(lsp): Use find references to display clone locations
- 0dc897d7e7c9d9a92abbf6f7278f3fd9ce18754a Merge pull request #11 from salab/intellij-changelog-update-0.5.1
- 2a815d61374a2fd01e3b5229f364d9af19eb07ac IntelliJ Changelog Update - 0.5.1

## [0.5.1] - 2024-10-24

- 471fc1c4e430ef45f4bc0a1241c19cc506d12596 v0.5.1
- 3ef3efa31129ccc720e1520aa58610a2da6e11fc impr(lsp): Readable clone locations
- 1f51f9be29dbc164a6e231a91803fff3b1076a49 Merge pull request #10 from salab/intellij-changelog-update-0.5.0
- 3cb8261de6ecfdbda9da8d31350d3daa632ff6a1 IntelliJ Changelog Update - 0.5.0

## [0.5.0] - 2024-10-21

- 7e5f1e5ac9a8bf89d1d56da0bad30ee31d1427f3 v0.5.0
- fe62e2b8efecb11960d1acd30b7f8e833be78ac6 feat: Display info of clone set even if all clones are edited
- 36f2ad3c0f8aefba28e6c9f6c023523e61e39bbc feat: Use 'after' tree to search for clones
- e2f97f014c078c709f91c978becdb3b2d9ec6f8a perf: Use ttl for cache
- 4e0071f25656faf4fb2e89ee692e1f5a6ba60761 perf: Use debounce of 0.5s for LSP server
- 67a4ffd4147980a8f4571ec9c1f3dc6989a58fbb perf: Implement calculation cache for LSP server
- 0ccbea5daa43cef3e843c1c74f3c78fd9aafe438 perf: Defer calculation of binary / contents
- f2e0db2bd3e160bb5fe32e4b7fb043ac574eea6f perf: Skip reading file entirely if binary
- f48d255df14654593b3954382174d0ddaa962a0c feat: Implement rate limiter for use with LSP server
- e15db3c2bb54d93525a77f216f48088242d748a0 fix(lsp): Run deduplication was not working
- 6d930f4349ce1be88b797c949da743c2374377b3 Merge pull request #9 from salab/intellij-changelog-update-0.4.2
- 451db79ffdb89659cdb66debb6f7ec68b4d8613d IntelliJ Changelog Update - 0.4.2

## [0.4.2] - 2024-09-21

- 74920c124634ec41a7c4b720762c0764c1da2bf7 v0.4.2
- 32e09f8eb79ff61743c4f446fd8045a72491abe6 Remove specification of plugin until build
- e1994df0d39309bf622607731de35218d94e5fe2 Update README
- 0703e63e956f909beff5e7fea28dd1004100a4e3 Merge pull request #8 from salab/intellij-changelog-update-0.4.1
- 3e5f91b5b67efe76dae7b2ce07108227a18ee9df IntelliJ Changelog Update - 0.4.1

## [0.4.1] - 2024-08-14

- b5025a042c6046e04a8843e3db837c5afa450e34 v0.4.1
- dbe952c0be867258448cfbaf5c0b66ffba6fcb27 fix: Skip detection of submodules entirely
- ea543e34c4d395fe0ed417699dcaf3fbf77d68c4 Merge pull request #7 from salab/intellij-changelog-update-0.4.0
- 1ee0a7e6fbb551ea83dceb14fbbb22342724323b IntelliJ Changelog Update - 0.4.0

## [0.4.0] - 2024-08-08

- c418d3f51c576709ff0072bc9bd2244269cb4bfe v0.4.0
- 6905facfccb81f5ea67d784d9d9eb96afe5d9850 fix(vscode): Ensure exec perm on the file
- 0667dd8ffe71e1c57b51f82a1fc4f973de1c8311 fix(lsp): Fix concurrent map access crash
- c026655a66dc957914f3dc83e2587bd104ca66ee feat: Add timeout to searching
- 695a918bc38285d9c785144394605787153772d2 Merge pull request #6 from salab/intellij-changelog-update-0.3.12
- 2656730321d35abfba6dcac13f0268d63450b769 IntelliJ Changelog Update - 0.3.12

## [0.3.12] - 2024-07-18

- 6a1ad044ef50b662ea7c453bed36e8332820ef1d v0.3.12
- c3163e525a5cc9838c928318def818c3ca91a53c Fix: fleccs detection was not working when comparing range containing empty lines
- a21b7f7dad0c4f6c9a995fdac3dd9fc62db9d4b1 Fix: ignore gitignore-d files/dirs before comparison
- c340b79e8799e034e025fd731c67b3bf8261b496 Fix: display errors
- ecad9b466ed30377d94d4a5ac725fa639c3d2b7a Merge pull request #5 from salab/intellij-changelog-update-0.3.11
- f9f98564e7105c7acbc7debb98ee536e72d2d73f IntelliJ Changelog Update - 0.3.11

## [0.3.11] - 2024-07-11

- f18de46 v0.3.11
- 17949e3 fix(intellij-plugin): download location was invalid on some platforms
- b50c0f1 Merge pull request #4 from salab/intellij-changelog-update-0.3.10
- 0dd1f85 IntelliJ Changelog Update - 0.3.10

## [0.3.10] - 2024-07-11

- ec69eb2 v0.3.10
- b1b88e3 fix(intellij-plugin): add missing arch switches
- 435cd49 Merge pull request #3 from salab/intellij-changelog-update-0.3.9
- 1ef490c IntelliJ Changelog Update - 0.3.9

## [0.3.9] - 2024-07-10

- 71fc4b8 0.3.9
- 062ddfa Add short description of the plugins
- 232da9c Merge pull request #2 from salab/intellij-changelog-update-0.3.8
- 57066ad IntelliJ Changelog Update - 0.3.8

## [0.3.8] - 2024-07-04

- a8b28f4 v0.3.8
- 906dfbf Remove default icon per request
- 1575422 Merge pull request #1 from salab/intellij-changelog-update-0.3.7
- aceee4d IntelliJ Changelog Update - 0.3.7

## [0.3.7] - 2024-07-04

- 5fb7679 v0.3.7
- 0b54441 Merge action to be triggered automatically
- 9c08dc4 Revert "Fix trigger"

## [0.3.3] - 2024-07-04

- First release of IntelliJ plugin of ICCheck

[Unreleased]: https://github.com/salab/iccheck/compare/v0.8.0...HEAD
[0.8.0]: https://github.com/salab/iccheck/compare/v0.7.6...v0.8.0
[0.7.6]: https://github.com/salab/iccheck/compare/v0.7.5...v0.7.6
[0.7.5]: https://github.com/salab/iccheck/compare/v0.7.4...v0.7.5
[0.7.4]: https://github.com/salab/iccheck/compare/v0.7.3...v0.7.4
[0.7.3]: https://github.com/salab/iccheck/compare/v0.7.2...v0.7.3
[0.7.2]: https://github.com/salab/iccheck/compare/v0.7.1...v0.7.2
[0.7.1]: https://github.com/salab/iccheck/compare/v0.7.0...v0.7.1
[0.7.0]: https://github.com/salab/iccheck/compare/v0.6.0...v0.7.0
[0.6.0]: https://github.com/salab/iccheck/compare/v0.5.1...v0.6.0
[0.5.1]: https://github.com/salab/iccheck/compare/v0.5.0...v0.5.1
[0.5.0]: https://github.com/salab/iccheck/compare/v0.4.2...v0.5.0
[0.4.2]: https://github.com/salab/iccheck/compare/v0.4.1...v0.4.2
[0.4.1]: https://github.com/salab/iccheck/compare/v0.4.0...v0.4.1
[0.4.0]: https://github.com/salab/iccheck/compare/v0.3.12...v0.4.0
[0.3.12]: https://github.com/salab/iccheck/compare/v0.3.11...v0.3.12
[0.3.11]: https://github.com/salab/iccheck/compare/v0.3.10...v0.3.11
[0.3.10]: https://github.com/salab/iccheck/compare/v0.3.9...v0.3.10
[0.3.9]: https://github.com/salab/iccheck/compare/v0.3.8...v0.3.9
[0.3.8]: https://github.com/salab/iccheck/compare/v0.3.7...v0.3.8
[0.3.7]: https://github.com/salab/iccheck/compare/v0.3.3...v0.3.7
[0.3.3]: https://github.com/salab/iccheck/commits
