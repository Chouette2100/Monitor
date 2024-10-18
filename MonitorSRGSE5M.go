package main

import (
	"fmt"
	"log"
	"time"

	"github.com/Chouette2100/srdblib"
)

type Result struct {
	Eventid string
	Ct      int
	DT      bool
}

func MonitorSRGSE5M(
	ch chan<- string,
	it int,
	sm int, // 毎時 sm分からit分おきにデータを取得する
) {

	var err error

	resmap := make(map[string]*[2]int)

	for {
		tnow := time.Now()
		tstart := tnow.Add(time.Duration(-sm) * time.Minute).Truncate(time.Duration(it) * time.Minute).Add(time.Duration(sm) * time.Minute)
		tend := tstart.Add(time.Duration(it) * time.Minute)
		wait := time.Until(tend)
		log.Printf("tstart %s tend %s wait %+v\n", tstart.Format("2006-01-02 15:04:05"), tend.Format("2006-01-02 15:04:05"), wait)
		//	待機
		time.Sleep(wait)

		//	it分間の間の取得データ数を調べる
		sqlst := "	select eventid, count(*) ct from points "
		sqlst += "		where ts between ? and ? "
		sqlst += "  and eventid in (select eventid from event where now() between starttime and endtime order by eventid)  "
		sqlst += "      group by eventid order by eventid "

		var rows []interface{}
		var result Result
		rows, err = srdblib.Dbmap.Select(result, sqlst, tstart, tend)
		if err != nil {
			err = fmt.Errorf("Dbmap.Select() error %v", err)
			log.Printf("%s\n", err)
			ch <- err.Error()
			return
		}
		for rm := range resmap {
			resmap[rm][1] = 0
		}
		for _, row := range rows {
			r := row.(*Result)
			eventid := r.Eventid
			//	log.Printf("eventid %s  ct %d", eventid, r.Ct)
			if ct, ok := resmap[eventid]; ok {
				if r.Ct != ct[0] {
					if r.Ct > ct[0] {
						log.Printf("eventid %s  ct %d -> %d\n", eventid, ct[0], r.Ct)
					} else {
						log.Printf("\x1b[31meventid %s  ct %d -> %d\x1b[39m\n", eventid, ct[0], r.Ct)
					}
					//
				}
				resmap[eventid][0] = r.Ct
				resmap[eventid][1] = 1
			} else {
				log.Printf("eventid %s  ct %d\n", eventid, r.Ct)
				resmap[eventid] = &[2]int{r.Ct, 1}
			}
		}
		for rm := range resmap {
			if resmap[rm][1] == 0 {
				if resmap[rm][0] != 0 {
					log.Printf("\x1b[31meventid %s  ct %d -> 0\x1b[39m\n", rm, resmap[rm][0])
				}
				resmap[rm][0] = 0
			}
		}
	}

}
