
extern crate time;

mod collector;
mod db;

use std::time::Duration;
use std::thread;
use collector::grab_data;
use collector::Company;

fn log_info(msg: &str) {
    let now = time::now();
    println!("INFO:{} -{}",now.asctime(), msg);
}

fn log_error(msg: &str) {
    println!("ERROR:{} - {}", time::now().asctime(), msg);
}

fn main() {

    //loop {
        let mut data : Vec<Company>;
        data = Vec::new();
        grab_data(&mut data).unwrap();
        match db::connect_to_db() {
            Ok(conn) => {
                db::insert_into_db(data, &conn);
                log_info("Inserted companies");
            },
            Err(_) => {
                log_error("Unable to connect to db");
            },
        }
        //thread::sleep(Duration::from_secs(3600));
    //}
}
