Tracepoint Why
sched/sched_process_exec Process launches — use this, not sys_enter_execve. It fires after the exec succeeds, so you get the real comm and PID.
syscalls/sys_enter_openat File opens — the richest signal, but also the noisiest.
sched/sched_process_exit Process deaths.
syscalls/sys_enter_connect Outbound TCP connections.
