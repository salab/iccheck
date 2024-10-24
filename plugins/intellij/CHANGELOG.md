<!-- Keep a Changelog guide -> https://keepachangelog.com -->

# ICCheck ChangeLog

## [Unreleased]

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

[Unreleased]: https://github.com/salab/iccheck/compare/v0.5.1...HEAD
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
