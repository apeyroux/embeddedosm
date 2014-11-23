package main

// embeddedosm -maxlat 48.8372 -maxlong 2.3111 -minlat 48.8028 -minlong 2.2501

import (
	"flag"
	"fmt"
	"github.com/j4/gosm"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strconv"
)

var (
	flmaxlat  = flag.Float64("maxlat", 0.0, "max latitude")
	flminlat  = flag.Float64("minlat", 0.0, "min latitude")
	flmaxlong = flag.Float64("maxlong", 0.0, "max longitude")
	flminlong = flag.Float64("minlong", 0.0, "min longitude")
	flinitdb  = flag.Bool("initdb", false, "init db")
)

const (
	VERSION = "1.0"
)

func getTileFromOSM(z int, x int, y int) ([]byte, error) {
	// utiliser le proxy http si il est config
	url := fmt.Sprintf("http://a.tile.openstreetmap.org/%d/%d/%d.png", z, x, y)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("Trouve pas la tuile sur WWW")
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Je ne comprend pas très bien le contenu de cette tuile")
	}

	log.Printf("hitwww z:%d x:%d y:%d from %s", z, x, y, url)

	return body, nil
}

func getsrvosm() string {
	srvs := []string{"a.tile.openstreetmap.org"}
	return srvs[rand.Intn(len(srvs))]
}

func mainhandler(w http.ResponseWriter, r *http.Request) {
	///	began := time.Now()
	path := r.URL.Path[1:]
	re := regexp.MustCompile("^([0-9]+)/([0-9]+)/([0-9]+).png$")
	tilecoord := re.FindStringSubmatch(path)

	if tilecoord != nil {
		// TODO: mettre du test sur les err de strconv
		z, _ := strconv.Atoi(tilecoord[1])
		x, _ := strconv.Atoi(tilecoord[2])
		y, _ := strconv.Atoi(tilecoord[3])
		//urlTile := fmt.Sprintf("http://%s/%d/%d/%d.png", getsrvosm(), z, x, y)

		log.Printf("demande de z:%d x:%d y:%d", z, x, y)
		tilePath := fmt.Sprintf("./tiles/%d-%d-%d.png", z, x, y)
		data, _ := ioutil.ReadFile(tilePath)
		w.Write(data)
		//			sql := fmt.Sprintf("SELECT data,* FROM tiles WHERE z=%d and x=%d and y=%d;", z, x, y)
		//QUERY_TILE_EXIST           = "SELECT data,dthr FROM tiles WHERE z=$1 AND x=$2 AND y=$3"
		// log.Printf(sql)
		// row := make(sqlite3.RowMap)
		// for s, err := db.Query(sql); err == nil; err = s.Next() {
		// 	var data []byte
		// 	err = s.Scan(&data, row) // Assigns 1st column to rowid, the rest to row
		// 	if err != nil {
		// 		log.Fatalf("err:%s", err)
		// 	}
		// 	//				log.Printf("d:%x\n", data)
		// 	w.Write(data)
		// }

	} else {
		http.NotFound(w, r)
	}
}

func osmhandler(w http.ResponseWriter, r *http.Request) {
	html, _ := ioutil.ReadFile("./index.html")

	w.Write(html)
}

func initCache() {
	for _, z := range []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19} {
		//for _, z := range []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14} {
		tMax := gosm.NewTileWithLatLong(*flmaxlat, *flmaxlong, z)
		tMin := gosm.NewTileWithLatLong(*flminlat, *flminlong, z)
		log.Printf("(%d-%d)+(%d-%d)", tMax.X, tMin.X, tMax.Y, tMin.Y)
		nbtiles := math.Abs((float64(tMax.X))-float64(tMin.X)) + math.Abs(float64(tMax.Y)-float64(tMin.Y))
		log.Printf("z:%d download %d tiles ...", tMax.Z, int64(nbtiles))

		for x := tMin.X; x <= tMax.X; x++ {
			for y := tMax.Y; y <= tMin.Y; y++ {
				tilePath := fmt.Sprintf("./tiles/%d-%d-%d.png", z, x, y)
				log.Printf("dwl a.tile.openstreetmap.org/%d/%d/%d.png", z, x, y)
				data, _ := getTileFromOSM(z, x, y)
				ioutil.WriteFile(tilePath, data, 0644)
				// args := sqlite3.NamedArgs{"$z": z, "$x": x, "$y": y, "$data": data}
				// err := c.Exec("INSERT INTO tiles VALUES($z, $x, $y, $data)", args) // $c will be NULL
				// if err != nil {
				// 	log.Printf("err: %s", err)
				// }
			}
		}
	}
}

func main() {
	flag.Parse()

	if *flinitdb {
		initCache()
		log.Printf("Initialisation de la base")
		os.Exit(0)
	}

	http.HandleFunc("/", mainhandler)
	http.Handle("/tiles/", http.FileServer(http.Dir("./tiles/")))
	http.HandleFunc("/osm/", osmhandler)
	http.ListenAndServe(":8080", nil)

}
