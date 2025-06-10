# Phase 3 Integration Testing Context

## Session Summary (January 11, 2025)

### Starting Point
- Began with Phase 2 complete, all unit tests passing
- No integration tests existed
- Project at 65% completion

### Ending Point  
- Phase 3 in progress with comprehensive integration test framework
- 13 out of 18 integration tests passing
- Project at 75% completion

## What Was Done

### Integration Test Framework
1. Created comprehensive test framework in `test/integration/framework_test.go`
   - Mock CallToolRequest implementation
   - Helper methods for all 9 MCP tools
   - WaitForContent helper for async testing
   - Cleanup method to remove all sessions

2. Tool Tests (`test/integration/tools_test.go`)
   - TestLaunchApp ✅
   - TestViewScreen ✅
   - TestSendKeys ✅
   - TestGetCursorPosition ✅
   - TestGetScreenSize ✅
   - TestResizeTerminal ✅ (fixed parameter type issue)
   - TestStopApp ✅
   - TestRestartApp ❌ (session becomes inactive)
   - TestListSessions ✅
   - TestConcurrentSessions ✅
   - TestSpecialKeys ✅
   - TestErrorHandling ✅
   - TestAnsiOutput ❌ (raw format not preserving ANSI)

3. Test App Tests (`test/integration/testapps_test.go`)
   - TestEchoApp ✅
   - TestMenuApp ❌ (timeout)
   - TestProgressApp ❌ (timeout)
   - TestAnsiFormatShowsCursor ✅
   - TestScrollbackFormat ❌ (not detecting historical content)

### Key Fixes Made

1. **Parameter Type Handling**:
   ```go
   // ResizeTerminal now accepts both int and float64
   var width float64
   if w, ok := args["width"].(float64); ok {
       width = w
   } else if w, ok := args["width"].(int); ok {
       width = float64(w)
   }
   ```

2. **Session Lifecycle**:
   - Added debug logging for argument extraction
   - Attempted fix for restart with sleep after start
   - Sessions need to stay alive for testing (use `sh -c` with sleep)

3. **Test Stability**:
   - Changed from bare commands to `sh -c "command; sleep n"`
   - This prevents immediate session termination
   - Built test apps are now executable binaries

### Integration Test Results
- **Passing**: 13 out of 18 tests
- **Failing**: 5 tests need fixes

## Critical Issues to Fix

### 1. Session Restart (TestRestartApp)
**Problem**: After restart, session state becomes inactive immediately
**Root Cause**: The readLoop goroutine lifecycle isn't properly managed
**Solution Needed**:
- Track readLoop goroutine with sync.WaitGroup or channel
- Ensure old readLoop is fully stopped before starting new one
- Consider adding a "restarting" state to prevent race conditions

### 2. Scrollback Format (TestScrollbackFormat)
**Problem**: Test expects more lines in scrollback than in plain view
**Root Cause**: Scrollback might not be adding lines properly for short-lived commands
**Solution Needed**:
- Debug scrollback addition in ScrollUp
- Ensure lines are added even if process exits quickly
- May need to adjust test to generate more output

### 3. ANSI Output (TestAnsiOutput)
**Problem**: Raw format not preserving ANSI escape sequences
**Root Cause**: The echo command might be stripping ANSI or shell interpreting it
**Solution Needed**:
- Use printf instead of echo -e
- Test with a program that definitely outputs ANSI (like the test apps)
- Verify raw format includes all escape sequences

### 4. Test App Timeouts
**Problem**: Menu and Progress apps timing out after 5 seconds
**Root Cause**: Test apps might be failing to start or communicate
**Solution Needed**:
- Add more logging to understand why apps aren't responding
- Check if apps are actually running
- May need to adjust test timeouts or app startup sequence

## Code Patterns Established

### Testing Patterns
```go
// Launch app and keep alive
sessionID := tf.LaunchApp("sh", []string{"-c", "command; sleep 1"})

// Wait for content with timeout
if !tf.WaitForContent(sessionID, "expected", 2*time.Second) {
    content := tf.ViewScreen(sessionID, "plain")
    t.Fatalf("Expected 'expected' but got: %s", content)
}

// Cleanup
defer tf.Cleanup()
```

### Parameter Handling Pattern
```go
// Handle multiple types for parameters
var param float64
if p, ok := args["param"].(float64); ok {
    param = p
} else if p, ok := args["param"].(int); ok {
    param = float64(p)
} else {
    // Handle error
}
```

## Next Session Priority

1. **Fix TestRestartApp** - Most critical for reliability
2. **Fix ANSI preservation** - Important for accurate terminal emulation
3. **Debug test app timeouts** - Needed for full test coverage
4. **Fix scrollback test** - Lower priority but needed for completeness
5. **Continue with error recovery implementation** - Main Phase 3 goal