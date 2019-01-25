
extern crate postgres;

use self::postgres::{Connection, SslMode};
use self::postgres::error::ConnectError;
use self::postgres::rows;

use collector::Company;
use std::error;
use std::collections::HashMap;

pub fn log_unwrap<T,E: error::Error>(res : Result<T,E>){
    match res {
        Ok(_) => {},
        Err(err) => println!("{:?}",err.description()),
    }
}


pub fn connect_to_db() -> Result<Connection,ConnectError> {
    let conn_res = Connection::connect("postgres://postgres:12345@localhost:5432/stocks",
    SslMode::None);
    match conn_res.as_ref() {
        Ok(conn) => {
            log_unwrap(conn.execute("CREATE TABLE IF NOT EXISTS companies(
                company_id      SERIAL PRIMARY KEY,
                company_name    text,
                type            text
            )", &[]));

            log_unwrap(conn.execute("CREATE TABLE IF NOT EXISTS price(
                company_id      integer,
                time            timestamp without time zone,
                price           double precision
            )",&[]));
        },
        Err(_) => {},
    }
    conn_res
}

pub fn insert_into_db(data : Vec<Company>, db: & Connection) {
    let name_id_map = get_companies(&db);
    for company in data {
        match name_id_map.get(&(company.name.trim().to_string())){
            Some(id) => log_unwrap(db.execute("Insert INTO price (company_id, time, price) VALUES ($1,now(),$2)",
                        &[&id,&company.price])),
            None => {
                log_unwrap(
                    db.execute("INSERT INTO
                    companies (company_name) VALUES ($1)", &[&company.name.trim()]));
            }, //should add company to db
        }
    }
}

fn get_companies(db: &Connection) -> HashMap<String,i32> {

    let mut name_id_map :HashMap<String,i32> = HashMap::new();
    let load_db = |companies : &mut rows::Rows| {
        for company in companies.into_iter() {
            let id:i32 = company.get(0);
            let name:String = company.get(1);
            name_id_map.insert(name,id);
        }
        name_id_map
    };
     db.query("SET timezone to 'Africa/Nairobi';",&[]).unwrap();
    match db.query("SELECT company_id, company_name FROM companies", &[]) {
        Ok(ref mut companies) => {
            return load_db(companies);
        },
        Err(err) => {
            panic!(err);
        },
    }
}
