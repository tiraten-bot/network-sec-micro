package main

import "time"

func maybeEject(rc RouteConfig, st *routeState, upstream string) {
    if rc.OutlierDetection == nil || !rc.OutlierDetection.Enabled { return }
    thr := rc.OutlierDetection.FailureThreshold
    if thr <= 0 { thr = 5 }
    if st.fail[upstream] >= thr {
        dur := time.Duration(rc.OutlierDetection.EjectDurationSec) * time.Second
        if dur <= 0 { dur = 30 * time.Second }
        st.ejectedUntil[upstream] = time.Now().Add(dur)
        st.fail[upstream] = 0
    }
}


