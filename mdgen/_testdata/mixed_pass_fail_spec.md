# Checkout

_File: `checkout.spec` · Time: 6.00s · Tags: regression_

## Summary

| Scope | Total | ✅ Passed | ❌ Failed | ⏭️ Skipped | Success rate |
| --- | --- | --- | --- | --- | --- |
| Scenarios | 2 | 1 | 1 | 0 | 50% |

## Scenarios

### ✅ Happy path — 00:00:01

#### Steps

- ✅ checkout completes _(00:00:01)_

### ❌ Bad path — 00:00:05

#### Steps

- ❌ checkout breaks _(00:00:00)_

  **Error:** `expected 200 got 500`

  <details><summary>Stack trace</summary>

  ```
  at handler.go:42
  at router.go:11
  ```

  </details>

