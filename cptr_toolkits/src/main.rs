extern crate sciter;

fn main() {
    let mut frame = sciter::Window::new();
    frame.load_file("minimal.htm");
    frame.run_app();
}