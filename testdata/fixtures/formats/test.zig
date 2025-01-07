//! top-level doc-comment

// normal comment

/// doc-comment
const Timestamp = struct {
    /// XXX: A comment in a struct
    nanos: u32,

    /// TODO: Ad function
    pub fn unixEpoch() Timestamp {}
};
