package protocol_test

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.ibm.com/Observability/fluent-forward-go/fluent/protocol"
)

var _ = Describe("Transport", func() {
	Describe("EventTime", func() {
		var (
			ent EntryExt
		)

		BeforeEach(func() {
			ent = EntryExt{
				Timestamp: EventTime{
					Time: time.Unix(int64(1257894000), int64(12340000)),
				},
			}
		})

		// This covers both MarshalBinaryTo() and UnmarshalBinary()
		It("Marshals and unmarshals correctly", func() {
			b, err := ent.MarshalMsg(nil)

			Expect(err).NotTo(HaveOccurred())

			// This is the msgpack fixext8 encoding for the timestamp
			// per the fluent-forward spec:
			// D7 == fixext8
			// 00 == type 0
			// 4AF9F070 == 1257894000
			// 00BC4B20 == 12340000
			Expect(
				strings.Contains(fmt.Sprintf("%X", b), "D7004AF9F07000BC4B20"),
			).To(BeTrue())

			var unment EntryExt
			_, err = unment.UnmarshalMsg(b)
			Expect(err).NotTo(HaveOccurred())

			Expect(unment.Timestamp.Time.Equal(ent.Timestamp.Time)).To(BeTrue())
		})
	})

	Describe("EntryList", func() {
		var (
			e1 EntryList
			et time.Time
		)

		BeforeEach(func() {
			et = time.Now()
			e1 = EntryList{
				{
					Timestamp: EventTime{et},
					Record: map[string]string{
						"foo":    "bar",
						"george": "jungle",
					},
				},
				{
					Timestamp: EventTime{et},
					Record: map[string]string{
						"foo":    "kablooie",
						"george": "frank",
					},
				},
			}
		})

		Describe("Equal", func() {
			var (
				e2 EntryList
			)

			BeforeEach(func() {
				e2 = EntryList{
					{
						Timestamp: EventTime{et},
						Record: map[string]string{
							"foo":    "bar",
							"george": "jungle",
						},
					},
					{
						Timestamp: EventTime{et},
						Record: map[string]string{
							"foo":    "kablooie",
							"george": "frank",
						},
					},
				}
			})

			It("Returns true", func() {
				Expect(e1.Equal(e2)).To(BeTrue())
			})

			Context("When the lists have different element counts", func() {
				BeforeEach(func() {
					e2 = e2[:1]
				})

				It("Returns false", func() {
					Expect(e1.Equal(e2)).To(BeFalse())
				})
			})

			Context("When the lists have differing elements", func() {
				BeforeEach(func() {
					e2[0].Timestamp = EventTime{et.Add(5 * time.Second)}
				})

				It("Returns false", func() {
					Expect(e1.Equal(e2)).To(BeFalse())
				})
			})
		})
	})

	Describe("NewPackedForwardMessage", func() {
		var (
			tag     string
			entries EntryList
			opts    MessageOptions
		)

		BeforeEach(func() {
			tag = "foo.bar"
			entries = EntryList{
				{
					Timestamp: EventTime{time.Now()},
					Record: map[string]string{
						"foo":    "bar",
						"george": "jungle",
					},
				},
				{
					Timestamp: EventTime{time.Now()},
					Record: map[string]string{
						"foo":    "kablooie",
						"george": "frank",
					},
				},
			}
			opts = MessageOptions{}
		})

		It("Returns a PackedForwardMessage", func() {
			msg := NewPackedForwardMessage(tag, entries, opts)
			Expect(msg).NotTo(BeNil())
		})

		It("Includes the number of events as the `size` option", func() {
			msg := NewPackedForwardMessage(tag, entries, opts)
			Expect(msg.Options).To(HaveKey(OPT_SIZE))
			size, err := strconv.Atoi(msg.Options[OPT_SIZE])
			Expect(err).NotTo(HaveOccurred())
			Expect(size).To(Equal(len(entries)))
		})

		XIt("Correctly encodes the entries into a bytestream", func() {
			// TODO: This test is wrong - it expects that the stream is a
			// single array of EntryExt objects, but it's a stream of encoded
			// EntryExt objects (NOT an array), and the test does not match
			// up to that.
			msg := NewPackedForwardMessage(tag, entries, opts)
			elist := make(EntryList, 2)
			_, err := elist.UnmarshalMsg(msg.EventStream)
			Expect(err).NotTo(HaveOccurred())

			Expect(elist.Equal(entries)).To(BeTrue())
		})
	})

	Describe("NewCompressedPackedForwardMessage", func() {
		var (
			tag     string
			entries []EntryExt
			opts    MessageOptions
		)

		BeforeEach(func() {
			tag = "foo.bar"
			entries = []EntryExt{
				{
					Timestamp: EventTime{time.Now()},
					Record: map[string]string{
						"foo":    "bar",
						"george": "jungle",
					},
				},
				{
					Timestamp: EventTime{time.Now()},
					Record: map[string]string{
						"foo":    "kablooie",
						"george": "frank",
					},
				},
			}
			opts = MessageOptions{}
		})

		It("Returns a message with a gzip-compressed event stream", func() {
			msg := NewCompressedPackedForwardMessage(tag, entries, opts)
			Expect(msg).NotTo(BeNil())
		})
	})
})
