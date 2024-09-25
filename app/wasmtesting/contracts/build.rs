use protobuf_codegen::{Codegen, Customize, CustomizeCallback};
use protoc_bin_vendored;
use protobuf::descriptor::field_descriptor_proto::Type;
use protobuf::reflect::FieldDescriptor;
use protobuf::reflect::MessageDescriptor;

fn main() {
    struct GenSerde;

    impl CustomizeCallback for GenSerde {
        fn message(&self, _message: &MessageDescriptor) -> Customize {
            Customize::default().before("#[derive(::serde::Serialize, ::serde::Deserialize)]")
        }

        fn field(&self, field: &FieldDescriptor) -> Customize {
            if field.proto().type_() == Type::TYPE_ENUM {
                // `EnumOrUnknown` is not a part of rust-protobuf, so external serializer is needed.
                Customize::default().before(
                    "#[serde(serialize_with = \"crate::serialize_enum_or_unknown\", deserialize_with = \"crate::deserialize_enum_or_unknown\")]")
            } else {
                Customize::default()
            }
        }

        fn special_field(&self, _message: &MessageDescriptor, _field: &str) -> Customize {
            Customize::default().before("#[serde(skip)]")
        }
    }

    Codegen::new()
    // Use `protoc` parser, optional.
    .protoc()
    // Use `protoc-bin-vendored` bundled protoc command, optional.
    .protoc_path(&protoc_bin_vendored::protoc_bin_path().unwrap())
    // All inputs and imports from the inputs must reside in `includes` directories.
    .includes(&["src/protos"])
    // Inputs must reside in some of include paths.
    .input("src/protos/connect.proto")
    // Specify output directory relative to Cargo output directory.
    .out_dir("src/")
    .customize_callback(GenSerde)
    .run_from_script();
}
