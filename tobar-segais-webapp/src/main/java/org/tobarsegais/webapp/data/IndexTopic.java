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

public class IndexTopic implements IndexChild {

    private static final long serialVersionUID = 1L;

    private final String title;
    private final String href;

    public IndexTopic(String title, String href) {
        this.title = title;
        this.href = href;
    }

    public static IndexTopic read(XMLStreamReader reader) throws XMLStreamException {
        if (reader.getEventType() != XMLStreamConstants.START_ELEMENT) {
            throw new IllegalStateException("Expecting a start element");
        }
        if (!"topic".equals(reader.getLocalName())) {
            throw new IllegalStateException("Expecting a <topic> element, got a <" + reader.getLocalName() + ">");
        }
        String title = reader.getAttributeValue(null, "title");
        if (title == null)
            title = reader.getAttributeValue(null, "label");
        String href = reader.getAttributeValue(null, "href");
        int depth = 0;
        while (reader.hasNext() && depth >= 0) {
            switch (reader.next()) {
                case XMLStreamConstants.START_ELEMENT:
                    depth++;
                    break;
                case XMLStreamConstants.END_ELEMENT:
                    depth--;
                    break;
            }
        }
        return new IndexTopic(title, href);
    }

    public void write(XMLStreamWriter writer) throws XMLStreamException {
        writer.writeStartElement("topic");
        writer.writeAttribute("title", getTitle());
        writer.writeAttribute("href", getHref());
        writer.writeEndElement();
    }

    public String getHref() {
        return href;
    }

    public String getTitle() {
        return title;
    }

    @Override
    public String toString() {
        final StringBuilder sb = new StringBuilder();
        sb.append("IndexTopic");
        sb.append("{title='").append(title).append('\'');
        sb.append(", href='").append(href).append('\'');
        sb.append('}');
        return sb.toString();
    }
}
