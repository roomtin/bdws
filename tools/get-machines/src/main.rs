use clap::Parser;
use rayon::prelude::*;
use std::fmt::Write as fmt_write;
use std::fs::File;
use std::io::{prelude::*, BufReader, Write};
use std::io::{Error, ErrorKind};
use std::process::Command;
use std::sync::{Arc, Mutex};

/// A small utility to aid in figuring out which JMU CS lab machines are
/// currently not being used by other students, takes a list
/// of machines to check and writes out a list of machines with no users currently logged in
#[derive(Parser, Debug)]
#[clap(about, long_about = None)]
struct Args {
    /// A file containing the hostnames of the machines to check, one hostname per line
    machine_list: String,

    /// Pass to have information about which users are currently logged in to which machines
    /// printed out at the end
    #[clap(short, long)]
    display_used_machines: bool,

    /// Multithreaded mode, seems to not work well on STU, best used on lab machines only
    #[clap(short, long)]
    multithreaded: bool,
}

fn main() {
    let args = Args::parse();

    //open master_machine_list and read in all the hostnames into a vector
    let mut master_machine_list = Vec::new();
    let file = File::open(&args.machine_list).unwrap();
    let buf_reader = BufReader::new(file);
    for line in buf_reader.lines() {
        let line = line.unwrap();
        master_machine_list.push(line);
    }

    let machine_info = query_manager(args.multithreaded, master_machine_list);

    //if the user wants to display the machines that are currently in use, print them out
    if args.display_used_machines {
        for used_machines in machine_info.1 {
            let splits = used_machines.0.split_whitespace().collect::<Vec<&str>>();
            let user = splits[0];
            let location = splits[splits.len() - 1];
            let machine = used_machines.1.split(".").collect::<Vec<&str>>()[0];
            if location.contains(":") {
                println!("{} is using {} locally", user, machine);
            } else {
                let mut location = location.to_string();
                location.remove(0);
                location.pop();
                println!("{} is using {} from {}", user, machine, location);
            }
        }
    }

    //write the free machines to a string named "free_machines", one machine per line
    let mut free_machines = String::new();
    for free_machine in machine_info.0 {
        write!(&mut free_machines, "{}\n", free_machine).unwrap();
    }
    //write free_machines to a file named "free_machines.txt"
    let mut file = File::create("free_machines.txt").unwrap();
    file.write_all(free_machines.as_bytes()).unwrap();

    println!("\nDone!");
}

/**
 * Handles the querying of the hostnames in single threaded mode, calls a parallel iterator on the hostnames in multithreaded modef
 */
fn query_manager(multi: bool, hostnames: Vec<String>) -> (Vec<String>, Vec<(String, String)>) {
    if multi {
//@@ Shared Memory Parallelism
//@ # Arcs and Mutexes
//@ A mutex in rust is actually a wrapper for a value that can be accessed by multiple threads.
//@ This is done by using the Arc<T> and Mutex<T> types.
//@ The Arc<T> type is an Atomic Reference Counted pointer, and the Mutex<T> type is a Mutex.
//@ Creating a sharable protected value is done as in the following block:
//@{
        let free_machines_lock: Arc<Mutex<Vec<String>>> = Arc::new(Mutex::new(Vec::new()));
        let used_machines_lock: Arc<Mutex<Vec<(String, String)>>> =
            Arc::new(Mutex::new(Vec::new()));
//@}
        //iterate over all the hostnames in parallel with rayon and call query on each hostname
        hostnames.par_iter().for_each(|host| {
            let q_result = query(&host);
            //if the query returned an error, print the error and continue
            if let Err(e) = &q_result {
                println!("Error: {}", e);
            }

            if let Ok(option) = &q_result {
                if let Some(user) = option {
                    //if the option is Some, the machine is in use, add the machine to the used_machines vector
//@ ## Locking, Unlocking and Writing
//@ Locking and unlocking a mutex is done as in the following blocks:
//@{
                    used_machines_lock
                        .lock()
                        .unwrap()
                        .push((user.to_string(), host.to_string()));
//@}
                } else {
                    //if the option is None, the machine is free, add the machine to the free_machines vector
//@{
                    free_machines_lock.lock().unwrap().push(host.to_string());
//@}
                }
            }
        });

//@ ## Returning to Unprotected Values
//@ If you wish to return a protected value to an unprotected one, try the following:
//@{
        let free = Arc::try_unwrap(free_machines_lock)
            .unwrap()
            .into_inner()
            .unwrap();
//@}
        let used = Arc::try_unwrap(used_machines_lock)
            .unwrap()
            .into_inner()
            .unwrap();

        return (free, used);
    } else {
        let mut free: Vec<String> = Vec::new();
        let mut used: Vec<(String, String)> = Vec::new();

        for hostname in hostnames {
//@@ Executing Commands
//@ ## std::process::Command
//@ The Command type is used to execute external commands.
//@{
            let mut ssh_cmd = String::new();
            write!(ssh_cmd, "ssh -o ConnectTimeout=10 {} \"who\"", hostname).unwrap();

            let output = Command::new("sh").arg("-c").arg(ssh_cmd).output();
//@}
//@ ##### Recovering Text from the Output
//@ The output of a command is a `std::process::Output` type. In order to get the text from the output, you can use the `std::process::Output::stdout` method. And to get the error text, you can use the `std::process::Output::stderr` method. To remove non printable characters from the output, you can use the `std::str::from_utf8_lossy` method.
            match output {
                Ok(output) => {
//@{
                    let stdout = String::from_utf8_lossy(&output.stdout);
                    let stderr = String::from_utf8_lossy(&output.stderr);
//@}
                    if stderr.len() > 0 {
                        if stderr.contains("No route to host") {
                            let host = hostname.split(".").collect::<Vec<&str>>()[0];
                            println!("Error: {} is offline.", host);
                        } else {
                            println!("Error: STDERR from {}: {}", &hostname, stderr);
                        }
                    }
                    if stdout.len() > 0 {
                        used.push((stdout.to_string(), hostname.to_string()));
                    }
                    if stdout.len() == 0 && stderr.len() == 0 {
                        free.push(hostname);
                    }
                }
                Err(error) => {
                    println!("Command error: {}", error);
                }
            }
        }
        return (free, used);
    }
}

/**
 * Queries a hostname and returns the output of the command if there was any
 * output. Otherwise, returns an error message. Most of this was just copied
 * directly from the above function. There's probably a more elegant way to
 * have both multi threaded and single threaded versions of the program, but
 * I didn't bother to work it out.
 */
fn query(hostname: &String) -> Result<Option<String>, Error> {
    let mut ssh_cmd = String::new();
    write!(ssh_cmd, "ssh -o ConnectTimeout=10 {} \"who\"", hostname).unwrap();

    let output = Command::new("sh").arg("-c").arg(ssh_cmd).output();

    match output {
        Ok(output) => {
            let stdout = String::from_utf8_lossy(&output.stdout);
            let stderr = String::from_utf8_lossy(&output.stderr);
            //if stderr is not empty then there was an error connecting to the machine
            if stderr.len() > 0 {
                //if the error is that the machine says "No route to host" then the machine is offline
                if stderr.contains("No route to host") {
                    let host = hostname.split(".").collect::<Vec<&str>>()[0];
                    let mut error_msg = String::new();
                    write!(&mut error_msg, "{} is offline.", host)
                        .expect("Failed to write to string");
                    return Err(Error::new(ErrorKind::Other, error_msg));
                // else the error is something abnormal. return just that error from stderr
                } else {
                    let mut error_msg = String::new();
                    write!(&mut error_msg, "STDERR from {}: {}", &hostname, stderr)
                        .expect("Failed to write to string");
                    return Err(Error::new(ErrorKind::Other, error_msg));
                }
            }
            //if stdout has a length greater than 0, then the machine is in use
            if stdout.len() > 0 {
                return Ok(Some(stdout.to_string()));
            } else {
                return Ok(None);
            }
        }
        //if there was an error running the command, return that error. This shouldn't happen.
        Err(error) => Err(error),
    }
}
