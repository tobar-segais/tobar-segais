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
import java.io.Serializable;

public class IndexSubpath implements Serializable {

    private static final long serialVersionUID = 1L;

    private final String keyword;

    public IndexSubpath(String keyword) {
        this.keyword = keyword;
    }

    public static IndexSubpath read(XMLStreamReader reader) throws XMLStreamException {
        if (reader.getEventType() != XMLStreamConstants.START_ELEMENT) {
            throw new IllegalStateException("Expecting a start element");
        }
        if (!"subpath".equals(reader.getLocalName())) {
            throw new IllegalStateException("Expecting a <subpath> element, got a <" + reader.getLocalName() + ">");
        }
        String keyword = reader.getAttributeValue(null, "keyword");
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
        return new IndexSubpath(keyword);
    }

    public void write(XMLStreamWriter writer) throws XMLStreamException {
        writer.writeStartElement("subpath");
        writer.writeAttribute("keyword", getKeyword());
        writer.writeEndElement();
    }

    public String getKeyword() {
        return keyword;
    }

    @Override
    public String toString() {
        final StringBuilder sb = new StringBuilder();
        sb.append("Subpath");
        sb.append("{keyword='").append(keyword).append('\'');
        sb.append('}');
        return sb.toString();
    }
}
