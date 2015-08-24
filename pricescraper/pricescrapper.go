package pricescraper

import (
	//"fmt"
	"golang.org/x/net/html"
	//"io/ioutil"
	"container/list"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

type Company struct {
	Name      string
	LastPrice float64
	CType     string
}

type elempred func(*html.Node) bool

func CheckError(e error) {
	if e != nil {
		panic(e)
	}
}

//depth first search for the node which passes the given test
func SearchForElement(root *html.Node, tester elempred) *html.Node {
	stack := list.New()
	currElem := root
	for currElem != nil {
		if tester(currElem) == true {
			return currElem
		}
		for e := currElem.FirstChild; e != nil; e = e.NextSibling {
			stack.PushFront(e)
		}
		topStack := stack.Front()
		currElem = topStack.Value.(*html.Node)
		stack.Remove(topStack)
	}
	return nil
}

func hasAttrVal(n *html.Node, attr string, val string) bool {
	for _, a := range n.Attr {
		if a.Key == attr && a.Val == val {
			return true
		}
	}
	return false
}

func hasAnyAttrVal(n *html.Node, key string, val map[string]bool) bool {
	for _, a := range n.Attr {
		if a.Key == key {
			if _, ok := val[a.Val]; ok {
				return true
			}
		}
	}
	return false
}

func GrabData() map[string]Company {
	resp, err := http.Get("https://www.nse.co.ke/market-statistics/equity-statistics.html?view=statistics")
	CheckError(err)
	defer resp.Body.Close()
	doc, err := html.Parse(resp.Body)
	CheckError(err)

	isStockTable := func(n *html.Node) bool {
		if n.Data == "table" {
			ret := hasAttrVal(n, "class", "marketStats table table-striped")
			return ret
		}
		return false
	}

	isStockRow := func(n *html.Node) bool {
		if n.Data == "tr" {
			ret := hasAnyAttrVal(n, "class",
				map[string]bool{"row0": true,
					"row1": true,
					"row2": true})
			return ret
		}
		return false
	}

	exp, err := regexp.Compile("[0-9]+")
	CheckError(err)
	sanitizeName := func(n string) string {
		index := exp.FindStringIndex(n)
		if index == nil {
			return n
		}
		return strings.Trim(n[:uint64(index[0])], " ")
	}

	companyTypes := map[string]bool{"AUTOMOBILES AND ACCESSORIES": true,
		"AGRICULTURAL":                     true,
		"BANKING":                          true,
		"COMMERCIAL AND SERVICES":          true,
		"CONSTRUCTION AND ALLIED":          true,
		"ENERGY AND PETROLEUM":             true,
		"INSURANCE":                        true,
		"INVESTMENT":                       true,
		"INVESTMENT SERVICES":              true,
		"MANUFACTURING AND ALLIED":         true,
		"TELECOMMUNICATION AND TECHNOLOGY": true}
	prices := make(map[string]Company)
	table := SearchForElement(doc, isStockTable)
	if table != nil {
		row := SearchForElement(table, isStockRow)
		if row != nil {
			row = row.Parent.FirstChild
		}
		companyName := ""
		companyType := ""
		for row != nil {
			for el := row.FirstChild; el != nil; el = el.NextSibling {
				if el.FirstChild != nil {
					if el.FirstChild.Data == "Company" {
						break
					} else if strings.Contains(el.FirstChild.Data, "Ltd") ||
						strings.Contains(el.FirstChild.Data, "Ord") {
						companyName = sanitizeName(el.FirstChild.Data)
					} else if _, ok := companyTypes[el.FirstChild.Data]; ok {
						companyType = el.FirstChild.Data
					} else {
						companyPrice, err := strconv.ParseFloat(strings.Trim(el.FirstChild.Data, " "), 32)
						if err == nil {
							prices[companyName] = Company{Name: companyName, CType: companyType, LastPrice: companyPrice}
						} else {
							companyPrice = -1.0
						}
					}
				}
			}
			row = row.NextSibling
		}
	}
	return prices
}
