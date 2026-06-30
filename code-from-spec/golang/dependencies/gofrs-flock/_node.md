# SPEC/golang/dependencies/gofrs-flock

Cross-platform file locking for Go:
`github.com/gofrs/flock`.

Thread-safe, supports shared and exclusive locks with
non-blocking attempts and context-based timeouts via
polling. BSD 3-Clause license.

# Public

## Import

```go
import "github.com/gofrs/flock"
```

## Creating a lock

```go
f := flock.New("/path/to/lock.file")
```

Creates a `*Flock` instance for the given path. The lock
file is created on first lock attempt if it does not
exist.

## Exclusive lock — non-blocking

```go
locked, err := f.TryLock()
if err != nil {
    return err
}
if !locked {
    // lock held by another process
}
defer f.Unlock()
```

`TryLock` attempts to acquire an exclusive lock without
blocking. Returns `(true, nil)` on success, `(false, nil)`
if the lock is held by another process.

## Exclusive lock — with timeout

```go
ctx, cancel := context.WithTimeout(
    context.Background(), 30*time.Second)
defer cancel()

locked, err := f.TryLockContext(ctx, 100*time.Millisecond)
if err != nil {
    return err // may be context.DeadlineExceeded
}
if !locked {
    // lock not acquired
}
defer f.Unlock()
```

`TryLockContext` polls with the given retry delay until
the lock is acquired, an error occurs, or the context
deadline expires. On timeout, returns
`(false, context.DeadlineExceeded)`.

## Shared lock — non-blocking

```go
locked, err := f.TryRLock()
```

Same semantics as `TryLock` but acquires a shared lock.
Multiple shared locks can coexist; an exclusive lock
blocks shared locks.

## Shared lock — with timeout

```go
locked, err := f.TryRLockContext(ctx, 100*time.Millisecond)
```

Same polling semantics as `TryLockContext` but for shared
locks.

## Unlock

```go
err := f.Unlock()
```

Releases the lock (exclusive or shared). Safe to call
if already unlocked (no-op). Does not delete the lock
file from disk.

## Status

```go
f.Locked()  // true if exclusive lock held
f.RLocked() // true if shared lock held
```
