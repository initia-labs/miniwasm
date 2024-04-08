pub struct Contract {}

impl Default for Contract {
  fn default() -> Self {
    Self::new()
  }
}

impl<'a> Contract {
  fn new() -> Self {
    Self {}
  }
}
