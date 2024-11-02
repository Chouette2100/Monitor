package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	//	"strings"
	//	"strconv"
	//	"time"

	//	"net/http"

	"github.com/go-gorp/gorp"
	//      "gopkg.in/gorp.v2"

	//	"github.com/dustin/go-humanize"

	//	"github.com/Chouette2100/exsrapi"
	//	"github.com/Chouette2100/srapi"
	"github.com/Chouette2100/srdblib"
	"github.com/Chouette2100/exsrapi"
)

/*



00AA00	新規作成
00AB00	チェックの対象としたイベントの総数を表示する
00AC00	獲得ポイントを取得したルームの総数を表示する

*/

const Version = "00AC00"



//	Monitor	daemon、cronの実行結果を監視する
func main() {

	var (
		interval    = flag.Int("interval", 5, "int flag")
		startminute    = flag.Int("startminute", 1, "int flag")
	)

	//	ログ出力を設定する
	logfile, err := exsrapi.CreateLogfile(Version, srdblib.Version)
	if err != nil {
		panic("cannnot open logfile: " + err.Error())
	}
	defer logfile.Close()
	//	log.SetOutput(logfile)
	log.SetOutput(io.MultiWriter(logfile, os.Stdout))

	flag.Parse()

	log.Printf("param -startminute : %d\n", *startminute)
	log.Printf("param -interval : %d\n", *interval)


	//	データベースとの接続をオープンする。
	var dbconfig *srdblib.DBConfig
	dbconfig, err = srdblib.OpenDb("DBConfig.yml")
	if err != nil {
		err = fmt.Errorf("srdblib.OpenDb() returned error. %w", err)
		log.Printf("%s\n", err.Error())
		return
	}
	if dbconfig.UseSSH {
		defer srdblib.Dialer.Close()
	}
	defer srdblib.Db.Close()

	log.Printf("********** Dbhost=<%s> Dbname = <%s> Dbuser = <%s> Dbpw = <%s>\n",
		(*dbconfig).DBhost, (*dbconfig).DBname, (*dbconfig).DBuser, (*dbconfig).DBpswd)

	//	gorpの初期設定を行う
	dial := gorp.MySQLDialect{Engine: "InnoDB", Encoding: "utf8mb4"}
	srdblib.Dbmap = &gorp.DbMap{Db: srdblib.Db, Dialect: dial, ExpandSliceArgs: true}

	srdblib.Dbmap.AddTableWithName(srdblib.User{}, "user").SetKeys(false, "Userno")
	srdblib.Dbmap.AddTableWithName(srdblib.Userhistory{}, "userhistory").SetKeys(false, "Userno", "Ts")

	/*
	//      cookiejarがセットされたHTTPクライアントを作る
	client, jar, err := exsrapi.CreateNewClient("anonymous")
	if err != nil {
		err = fmt.Errorf("CreateNewClient() returned error. %w", err)
		log.Printf("%s\n", err.Error())
		return
	}
	//      すべての処理が終了したらcookiejarを保存する。
	defer jar.Save() //	忘れずに！
	*/

	ch := make(chan string)

	go MonitorSRGSE5M(ch, *interval, *startminute)

	//	終了待ち
	//	select {
	//	}
	msg := <-ch
	log.Printf("MonitorSRGSE5M() returned <%s>\n", msg)
}



