/* Copyright 2021 Victor Penso

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>. */

package main

import (
        "io/ioutil"
        "os/exec"
        "log"
        "strings"
        "strconv"
        "github.com/prometheus/client_golang/prometheus"
)

func SShareData() []byte {
        cmd := exec.Command( "sshare", "-n", "-P", "-o", "account,NormUsage,NormShares,LevelFS" )
        stdout, err := cmd.StdoutPipe()
        if err != nil {
                log.Fatal(err)
        }
        if err := cmd.Start(); err != nil {
                log.Fatal(err)
        }
        out, _ := ioutil.ReadAll(stdout)
        if err := cmd.Wait(); err != nil {
                log.Fatal(err)
        }
        return out
}

type SShareMetrics struct {
        normusage float64
		normshares float64
		levelfs float64
}

func ParseSShareMetrics() map[string]*SShareMetrics {
        accounts := make(map[string]*SShareMetrics)
        lines := strings.Split(string(SShareData()), "\n")
        for _, line := range lines {
                if ! strings.HasPrefix(line,"  ") {
                        if strings.Contains(line,"|") {
                                account := strings.Trim(strings.Split(line,"|")[0]," ")
                                _,key := accounts[account]
                                normusage,_ := strconv.ParseFloat(strings.Split(line,"|")[1],64)
								normshares,_ := strconv.ParseFloat(strings.Split(line,"|")[2],64)
								levelfs,_ := strconv.ParseFloat(strings.Split(line,"|")[3],64)
								if normusage > 0 && normshares > 0 {
									// only show active accounts
									if !key {
										accounts[account] = &SShareMetrics{0,0,0}
									}
									accounts[account].normusage = normusage
									accounts[account].normshares = normshares
									accounts[account].levelfs = levelfs
								}
                        }
                }
        }
        return accounts
}

type SShareCollector struct {
        normusage *prometheus.Desc
		normshares *prometheus.Desc
		levelfs *prometheus.Desc
}

func NewSShareCollector() *SShareCollector {
        labels := []string{"account"}
        return &SShareCollector{
                normusage: prometheus.NewDesc("slurm_account_normusage","NormUsage for account" , labels,nil),
				normshares: prometheus.NewDesc("slurm_account_normshares","NormShares for account" , labels,nil),
				levelfs: prometheus.NewDesc("slurm_account_levelfs","LevelFS for account" , labels,nil),
        }
}

func (fsc *SShareCollector) Describe(ch chan<- *prometheus.Desc) {
        ch <- fsc.normusage
		ch <- fsc.normshares
		ch <- fsc.levelfs
}

func (fsc *SShareCollector) Collect(ch chan<- prometheus.Metric) {
        fsm := ParseSShareMetrics()
        for f := range fsm {
                ch <- prometheus.MustNewConstMetric(fsc.normusage, prometheus.GaugeValue, fsm[f].normusage, f)
				ch <- prometheus.MustNewConstMetric(fsc.normshares, prometheus.GaugeValue, fsm[f].normshares, f)
				ch <- prometheus.MustNewConstMetric(fsc.levelfs, prometheus.GaugeValue, fsm[f].levelfs, f)
        }
}