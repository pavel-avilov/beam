/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package org.apache.beam.templates.transforms;

import java.util.List;
import java.util.Map;
import org.apache.beam.sdk.coders.AvroCoder;
import org.apache.beam.sdk.coders.NullableCoder;
import org.apache.beam.sdk.coders.StringUtf8Coder;
import org.apache.beam.sdk.io.gcp.pubsub.PubsubIO;
import org.apache.beam.sdk.io.gcp.pubsub.PubsubMessage;
import org.apache.beam.sdk.io.kafka.KafkaIO;
import org.apache.beam.sdk.transforms.MapElements;
import org.apache.beam.sdk.transforms.PTransform;
import org.apache.beam.sdk.values.KV;
import org.apache.beam.sdk.values.PBegin;
import org.apache.beam.sdk.values.PCollection;
import org.apache.beam.sdk.values.PDone;
import org.apache.beam.sdk.values.TypeDescriptor;
import org.apache.beam.templates.avro.TaxiRide;
import org.apache.beam.templates.avro.TaxiRidesKafkaAvroDeserializer;
import org.apache.beam.templates.options.KafkaToPubsubOptions;
import org.apache.beam.vendor.guava.v26_0_jre.com.google.common.base.Charsets;
import org.apache.beam.vendor.guava.v26_0_jre.com.google.common.collect.ImmutableMap;
import org.apache.kafka.common.serialization.StringDeserializer;

/** Different transformations over the processed data in the pipeline. */
public class FormatTransform {

  public enum FORMAT {
    PUBSUB,
    PLAINTEXT,
    AVRO;
  }

  /**
   * Configures Kafka consumer.
   *
   * @param bootstrapServers Kafka servers to read from
   * @param topicsList Kafka topics to read from
   * @param config configuration for the Kafka consumer
   * @return configured reading from Kafka
   */
  public static PTransform<PBegin, PCollection<KV<String, String>>> readFromKafka(
      String bootstrapServers, List<String> topicsList, Map<String, Object> config) {
    return KafkaIO.<String, String>read()
        .withBootstrapServers(bootstrapServers)
        .withTopics(topicsList)
        .withKeyDeserializerAndCoder(
            StringDeserializer.class, NullableCoder.of(StringUtf8Coder.of()))
        .withValueDeserializerAndCoder(
            StringDeserializer.class, NullableCoder.of(StringUtf8Coder.of()))
        .withConsumerConfigUpdates(config)
        .withoutMetadata();
  }

  /**
   * Configures Kafka consumer to read avros to {@link TaxiRide} format.
   *
   * @param bootstrapServers Kafka servers to read from
   * @param topicsList Kafka topics to read from
   * @param config configuration for the Kafka consumer
   * @return configured reading from Kafka
   */
  public static PTransform<PBegin, PCollection<KV<String, TaxiRide>>> readAvrosFromKafka(
      String bootstrapServers, List<String> topicsList, Map<String, Object> config) {
    return KafkaIO.<String, TaxiRide>read()
        .withBootstrapServers(bootstrapServers)
        .withTopics(topicsList)
        .withKeyDeserializerAndCoder(
            StringDeserializer.class, NullableCoder.of(StringUtf8Coder.of()))
        .withValueDeserializerAndCoder(
            TaxiRidesKafkaAvroDeserializer.class, AvroCoder.of(TaxiRide.class))
        .withConsumerConfigUpdates(config)
        .withoutMetadata();
  }

  /** Converts all strings into a chosen {@link FORMAT} and writes them into PubSub topic. */
  public static class FormatOutput extends PTransform<PCollection<String>, PDone> {

    private final KafkaToPubsubOptions options;

    public FormatOutput(KafkaToPubsubOptions options) {
      this.options = options;
    }

    @Override
    public PDone expand(PCollection<String> input) {
      if (options.getOutputFormat() == FORMAT.PUBSUB) {
        return input
            .apply(
                "convertMessagesToPubsubMessages",
                MapElements.into(TypeDescriptor.of(PubsubMessage.class))
                    .via(
                        (String json) ->
                            new PubsubMessage(json.getBytes(Charsets.UTF_8), ImmutableMap.of())))
            .apply(
                "writePubsubMessagesToPubSub",
                PubsubIO.writeMessages().to(options.getOutputTopic()));
      } else {
        return input.apply("writeToPubSub", PubsubIO.writeStrings().to(options.getOutputTopic()));
      }
    }
  }
}
