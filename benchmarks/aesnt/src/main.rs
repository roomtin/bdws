use aes::cipher::{generic_array::GenericArray, BlockDecrypt, KeyInit};
use aes::Aes128;
use chrono::{Timelike, Local};
use rayon::prelude::*;
use std::fmt::Write;
use std::io::Write as bWrite;
use std::mem::transmute;
use clap::Parser;



/// A conversion of a CS482 (Cryptography) program into a benchmark for the BDWS project
/// which will brute force any range of an AES key specified. This program is multi-threaded
/// and only needs to be started once per worker node. Furthermore, it tries to make use of the
/// Intel AES-NI instructions, so if run on an AMD processor, timing results may be inaccurate. 
#[derive(Parser, Debug)]
#[clap(author = "Author: Raleigh Martin", about)]
struct Args {
    //which key to start the guessing on, Must be a full u128 string
    start_guess: String,

    //which key to end the guessing on, Must be a full u128 string
    end_guess: String,
}
/**
 * Main function
 */
fn main() {
    //get args
    let args = Args::parse();

        //--------------------------
        //        AES DATA
        //--------------------------
        let iv = "A63319C14E9803288D56534C3F19CC81";
        let encrypted_text = String::from("ab304b07bd7bc9937a64022268b0d5a3");

        //Not used in the program other than to debug
        let _real_key = String::from("20342D96E60AE01CB32AFA9AF83C7239");

        //--------------------------
        //          SETUP
        //--------------------------

        //Convert the IV to an unsigned int
        let iv_num: u128 = u128::from_str_radix(iv, 16).expect("IV conversion failed.");

        //Tell user we started
        println!("Working...");

        let start_guess: u128 = u128::from_str_radix(&args.start_guess, 16).expect("Start guess conversion failed.");
        let end_guess: u128 = u128::from_str_radix(&args.end_guess, 16).expect("End guess conversion failed.");

        //Spawn a parallel iterator on the possible range
        //of keys and pass each one to a decrypt manager.
        //Parallelization should be handled automatically
        //on a per system basis.
        (start_guess..end_guess)
            .into_par_iter()
            .for_each(|key| {
                decrypt_manager(iv_num, key, &encrypted_text);

               //only turn on if debugging, slows things down.
               //prints out keys periodically to show progress.
               // if key % 0b100000000000000000000000 == 0 {
               //     println!("Passed 0x{:x}", key);
               // }
            });

        //Create clock for logging finish time
        let now = Local::now();
        println!("Done at: {:02}:{:02}", now.hour(), now.minute());
}

/**
 * Can handle long strings of ciphertext with the IV up front
 * Writes to a file when it finds the answer
 */
fn decrypt_manager(iv_num: u128, key_num: u128, bytes: &String) {
    //Main plaintext buffer
    let mut plaintext = String::new();

    //Unsafe code: memory transmutation from u128 to byte array
    //For this case, absolutely worth the simplicity
    //it offers. All inputs should be checked properly.
    let key_arr: [u8; 16] = unsafe { transmute(key_num.to_be()) };
    let mut iv: [u8; 16] = unsafe { transmute(iv_num.to_be()) };

    for i in (0..bytes.len()).step_by(32) {
        //Extract the current 16 bytes (block) to work on
        let current_bytes = &bytes[i..i + 32];

        //Convert the current block to an unsigned int
        let cur_block_num: u128 =
            u128::from_str_radix(current_bytes, 16).expect("Current Block conversion failed.");

        //Transmute the current block into a byte array
        let block_arr: [u8; 16] = unsafe { transmute(cur_block_num.to_be()) };

        //Decrypt the block
        let decrypted_block = block_decrypt(iv, key_arr, block_arr);

        //Set up IV for next decryption
        iv = block_arr;

        //Push the block's plaintext into the main buffer
        plaintext.push_str(decrypted_block.as_str());
    }

    //Validate correct answers. Write out a file if a match is found.
    if plaintext
        .chars()
        .all(|c| (c as u8) >= 32 && (c as u8) <= 126)
    {
        let now = Local::now(); //Nice that Chrono has local time
        let mut filename = String::new();
        write!(
            &mut filename,
            "results_from_{:02}{:02}_hours.txt",
            now.hour(), 
            now.minute(),
        )
        .expect("Couldn't write time");
        let mut file = std::fs::OpenOptions::new()
            .create(true)
            .append(true)
            .open(filename)
            .expect("Couldn't create new results file.");

        let mut output = String::new();
        write!(
            &mut output,
            "Plaintext: {}\n\nKey: {:x?}\n\n",
            plaintext, key_arr
        )
        .expect("Couldn't write to output buffer");
        file.write_all(output.as_bytes())
            .expect("Couldn't write to results file");
    }
}

/**
 * Decrypts a single block with Intel AES decryption instructions
 * if they're available, defaults to a software implementation on
 * non Intel CPUs. Uses the AES-NI library.
 */
fn block_decrypt(iv: [u8; 16], key_arr: [u8; 16], block_arr: [u8; 16]) -> String {
    let key = GenericArray::from_slice(&key_arr).to_owned();
    let mut block = GenericArray::from_slice(&block_arr).to_owned();

    //Create new cipher custom to the current key
    let mechanism = Aes128::new(&key);

    //Uses the Intel AES instructions to decrypt block
    //theoretically in a single
    mechanism.decrypt_block(&mut block);

    //Bitwise with IV
    //I think compiler will vectorize ?
    for i in 0..16 {
        block[i] = block[i] ^ iv[i];
    }

    //alternative to the above, probably a bit slower
    //block.iter_mut().zip(iv).for_each(|(byte, index)| *byte ^= index);

    //Convert decrypted Hex values to Text
    let mut textblock = vec![];
    block.iter().for_each(|x| textblock.push(*x as char));
    let text: String = textblock.iter().collect();

    //some fantasticly slick debugging here
    //dbg!(&text);
    //dbg!(&text.clone().chars().map(|c| c as u8).collect::<Vec<u8>>());

    text
}

//-------------------------------------------------------------------------------------------------------
//  Every function below this isn't for cracking AES, just debugging the original version of this program
//-------------------------------------------------------------------------------------------------------

// /**
// * Quick and dirty way to generate accurate keys
// * based on their description.
// */
//fn keygen() {
//    let mut start_key = String::new();
//    //push unknown part
//    for _i in 0..37 {
//        start_key.push_str("0");
//    }
//    //record length of unknown part
//    //push bumped 1
//    let starting_key_ones = start_key.len();
//    let key_num: u128 = u128::from_str_radix(start_key.as_str(), 2).expect("KeyGen failed keynum");
//    //push left over 1's
//    start_key.push_str("1");
//    start_key.push_str("1");
//    //push 0's
//    for _i in 0..87 {
//        start_key.push_str("0");
//    }
//    //push trailing 1's
//    start_key.push_str("11");
//    //printout
//    println!(
//        "Start Key String: {}\nStart Key Length: {}\nStarting Key Length Up To First 1's: {}",
//        &start_key,
//        start_key.len(),
//        starting_key_ones
//    );
//    println!("Start Guess Hex: {:0>32x}\n", key_num);
//    let key_num: u128 = u128::from_str_radix(start_key.as_str(), 2).expect("KeyGen failed keynum");
//    println!("Start Key Hex: {:0>32x}\n", key_num);
//    //---------------------------------------------------
//    let mut end_key = String::new();
//    //push unknown part
//    for _i in 0..37 {
//        end_key.push_str("1");
//    }
//    //push bumped 1
//    let key_num: u128 = u128::from_str_radix(end_key.as_str(), 2).expect("KeyGen failed keynum");
//    //push lefterover 1
//    end_key.push_str("1");
//    end_key.push_str("1");
//    //push 0's
//    for _i in 0..87 {
//        end_key.push_str("0");
//    }
//    //push trailing 1's
//    end_key.push_str("11");
//    //printout
//    println!(
//        "End Key String: {}\nEnd Key Length: {}",
//        &end_key,
//        end_key.len()
//    );
//    println!("End Guess Hex: {:0>32x}", key_num);
//    let key_num: u128 = u128::from_str_radix(end_key.as_str(), 2).expect("KeyGen failed keynum");
//    println!("End Key Hex: {:0>32x}", key_num);
//}
//
// /**
// * Debug mode for making sure program runs correctly
// * on a new system
// */
//fn debug() {
//    let test_iv: u128 = 0x9876543210FEDCBA9876543210FEDCBA;
//    let test_key: u128 = 0x000000000000000000000000000000FF;
//    let _str_block = String::from("03D735A237D13DEA619C0E810C6BC262");
//    let long_str_block =
//        String::from("11B2DC5005FA2C88D65C5DE3583E309B7CC3B1289FB7BC3A3433C1A9FE4FEBD8");
//    decrypt_manager(test_iv, test_key, &long_str_block);
//}