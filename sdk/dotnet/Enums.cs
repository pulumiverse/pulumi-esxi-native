// *** WARNING: this file was generated by pulumigen. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

using System;
using System.ComponentModel;
using Pulumi;

namespace Pulumiverse.EsxiNative
{
    [EnumType]
    public readonly struct BootFirmwareType : IEquatable<BootFirmwareType>
    {
        private readonly string _value;

        private BootFirmwareType(string value)
        {
            _value = value ?? throw new ArgumentNullException(nameof(value));
        }

        public static BootFirmwareType BIOS { get; } = new BootFirmwareType("bios");
        public static BootFirmwareType EFI { get; } = new BootFirmwareType("efi");

        public static bool operator ==(BootFirmwareType left, BootFirmwareType right) => left.Equals(right);
        public static bool operator !=(BootFirmwareType left, BootFirmwareType right) => !left.Equals(right);

        public static explicit operator string(BootFirmwareType value) => value._value;

        [EditorBrowsable(EditorBrowsableState.Never)]
        public override bool Equals(object? obj) => obj is BootFirmwareType other && Equals(other);
        public bool Equals(BootFirmwareType other) => string.Equals(_value, other._value, StringComparison.Ordinal);

        [EditorBrowsable(EditorBrowsableState.Never)]
        public override int GetHashCode() => _value?.GetHashCode() ?? 0;

        public override string ToString() => _value;
    }

    [EnumType]
    public readonly struct DiskType : IEquatable<DiskType>
    {
        private readonly string _value;

        private DiskType(string value)
        {
            _value = value ?? throw new ArgumentNullException(nameof(value));
        }

        public static DiskType Thin { get; } = new DiskType("thin");
        public static DiskType ZeroedThick { get; } = new DiskType("zeroedthick");
        public static DiskType EagerZeroedThick { get; } = new DiskType("eagerzeroedthick");

        public static bool operator ==(DiskType left, DiskType right) => left.Equals(right);
        public static bool operator !=(DiskType left, DiskType right) => !left.Equals(right);

        public static explicit operator string(DiskType value) => value._value;

        [EditorBrowsable(EditorBrowsableState.Never)]
        public override bool Equals(object? obj) => obj is DiskType other && Equals(other);
        public bool Equals(DiskType other) => string.Equals(_value, other._value, StringComparison.Ordinal);

        [EditorBrowsable(EditorBrowsableState.Never)]
        public override int GetHashCode() => _value?.GetHashCode() ?? 0;

        public override string ToString() => _value;
    }
}
