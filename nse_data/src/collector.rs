
extern crate hyper;
extern crate select;
extern crate regex;

use self::select::document::Document;
use self::select::predicate::{Class, Or,And};
use self::select::node::Node;


use self::hyper::Client;
use std::error::Error;
use std::io::Read;
use std::fmt;

pub struct Company {
    pub name  : String,
    pub price : f64
}

#[derive(Debug)]
pub enum DataParseErr {
    UnableToFetch,
}

impl Error for DataParseErr {
    fn description(&self) -> &str {
        match *self {
            DataParseErr::UnableToFetch => "Unable to fetch url"
        }
    }
}

impl fmt::Display for DataParseErr {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        match *self {
            DataParseErr::UnableToFetch => write!(f, "Unable to fetch data from url")
        }
    }
}

fn get_company(name :&Node, price :&Node) ->Option<Company> {
    let sanitize_name = |company_name : String|{
        let name_chars = company_name.chars();
        let mut clean_name = String::new();
        for ch in name_chars {
            if ch.is_numeric(){
                break;
            }
            clean_name.push(ch);
        }
        String::from(clean_name.trim())
    };
    match price.text().trim().parse(){
        Err(_) => return None,
        Ok(price_flt) => return Some(Company{
                name: sanitize_name(name.text()),
                price: price_flt
            }),
    };
}

pub fn grab_data(price_data :&mut Vec<Company>) ->Result<(),DataParseErr>{
    let url = "https://www.nse.co.ke/market-statistics/equity-statistics.html";
    let client = Client::new();

    let  res = client.get(url).send();
    let mut res_str = String::new();
    match res {
        Err(_) => return Err(DataParseErr::UnableToFetch),
        Ok(mut doc) => {
            match doc.read_to_string(&mut res_str) {
                Ok(_) => {},
                Err(_) => return Err(DataParseErr::UnableToFetch)
            };
        }
    };

    let document = Document::from_str(&res_str);

    let stock_predicate = And(And(Class("table"), Class("table-striped")), Class("marketStats"));
    match document.find(stock_predicate).first() {
        Some(stock_table) => grab_data_from_table(stock_table, price_data),
        None => { println!("didn't find table");
            return Ok(());
        }
    };

    return Ok(());
}

fn grab_data_from_table(stock_table :Node, price_data :&mut Vec<Company>){
    for row in stock_table.find(Or(Class("row0"), Class("row1"))).iter(){
        match row.find(Class("itemt")).first() {
            Some(name) => {
                match row.find(Class("tprice")).first() {
                    Some(price) =>{
                        match get_company(&name, &price) {
                            Some(acompany) => {
                                price_data.push(acompany);
                            },
                            None => continue
                        }
                    }
                    None => continue
                };
            }
            None => continue,
        };
    }
}
