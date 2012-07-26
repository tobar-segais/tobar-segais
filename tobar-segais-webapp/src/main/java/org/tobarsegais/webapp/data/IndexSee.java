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
import java.util.Collection;
import java.util.Collections;
import java.util.List;

public class IndexSee implements IndexChild {

    private static final long serialVersionUID = 1L;

    private final String keyword;
    private final List<IndexSubpath> children;

    public IndexSee(String keyword, Collection<IndexSubpath> children) {
        this.keyword = keyword;
        this.children = children == null || children.isEmpty()
                ? Collections.<IndexSubpath>emptyList()
                : Collections.unmodifiableList(new ArrayList<IndexSubpath>(children));
    }

    public static IndexSee read(XMLStreamReader reader) throws XMLStreamException {
        if (reader.getEventType() != XMLStreamConstants.START_ELEMENT) {
            throw new IllegalStateException("Expecting a start element");
        }
        if (!"see".equals(reader.getLocalName())) {
            throw new IllegalStateException("Expecting a <see> element, got a <" + reader.getLocalName() + ">");
        }
        String keyword = reader.getAttributeValue(null, "keyword");
        List<IndexSubpath> children = new ArrayList<IndexSubpath>();
        int depth = 0;
        while (reader.hasNext() && depth >= 0) {
            switch (reader.next()) {
                case XMLStreamConstants.START_ELEMENT:
                    if (depth == 0 && "subpath".equals(reader.getLocalName())) {
                        children.add(IndexSubpath.read(reader));
                    } else {
                        depth++;
                    }
                    break;
                case XMLStreamConstants.END_ELEMENT:
                    depth--;
                    break;
            }
        }
        return new IndexSee(keyword, children);
    }

    public void write(XMLStreamWriter writer) throws XMLStreamException {
        writer.writeStartElement("see");
        writer.writeAttribute("keyword", getKeyword());
        for (IndexSubpath topic : getChildren()) {
            topic.write(writer);
        }
        writer.writeEndElement();
    }

    public String getKeyword() {
        return keyword;
    }

    public List<IndexSubpath> getChildren() {
        return children == null ? Collections.<IndexSubpath>emptyList() : children;
    }

    @Override
    public String toString() {
        final StringBuilder sb = new StringBuilder();
        sb.append("IndexSee");
        sb.append("{keyword='").append(getKeyword()).append('\'');
        sb.append(", children=").append(getChildren());
        sb.append('}');
        return sb.toString();
    }

}
