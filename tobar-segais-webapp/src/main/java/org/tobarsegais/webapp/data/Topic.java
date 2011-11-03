/*
 * Copyright 2011 Stephen Connolly
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package org.tobarsegais.webapp.data;

import javax.xml.stream.XMLStreamConstants;
import javax.xml.stream.XMLStreamException;
import javax.xml.stream.XMLStreamReader;
import javax.xml.stream.XMLStreamWriter;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.Collection;
import java.util.List;

public class Topic extends Entry {

    private static final long serialVersionUID = 1L;

    public Topic(String label, String href, Topic... children) {
        this(label, href, Arrays.asList(children));
    }

    public Topic(String label, String href, Collection<Topic> children) {
        super(label, href, children);
    }

    public static Topic read(XMLStreamReader reader) throws XMLStreamException {
        if (reader.getEventType() != XMLStreamConstants.START_ELEMENT) {
            throw new IllegalStateException("Expecting a start element");
        }
        if (!"topic".equals(reader.getLocalName())) {
            throw new IllegalStateException("Expecting a <topic> element, got a <" + reader.getLocalName() + ">");
        }
        String label = reader.getAttributeValue(null, "label");
        String href = reader.getAttributeValue(null, "href");
        List<Topic> topics = new ArrayList<Topic>();
        int depth = 0;
        while (reader.hasNext() && depth >= 0) {
            switch (reader.next()) {
                case XMLStreamConstants.START_ELEMENT:
                    if (depth == 0 && "topic".equals(reader.getLocalName())) {
                        topics.add(Topic.read(reader));
                    } else {
                        depth++;
                    }
                    break;
                case XMLStreamConstants.END_ELEMENT:
                    depth--;
                    break;
            }
        }
        return new Topic(label, href, topics);
    }

    public void write(XMLStreamWriter writer) throws XMLStreamException {
        writer.writeStartElement("topic");
        writer.writeAttribute("label", getLabel());
        writer.writeAttribute("href", getHref());
        for (Topic topic : getChildren()) {
            topic.write(writer);
        }
        writer.writeEndElement();
    }

}
