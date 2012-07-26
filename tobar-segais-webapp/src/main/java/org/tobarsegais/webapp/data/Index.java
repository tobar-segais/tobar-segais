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

import org.apache.commons.io.IOUtils;

import javax.xml.stream.XMLInputFactory;
import javax.xml.stream.XMLStreamConstants;
import javax.xml.stream.XMLStreamException;
import javax.xml.stream.XMLStreamReader;
import javax.xml.stream.XMLStreamWriter;
import java.io.InputStream;
import java.io.Serializable;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.Collection;
import java.util.Collections;
import java.util.List;

public class Index implements Serializable {

    private static final long serialVersionUID = 1L;

    private final List<IndexEntry> children;

    public Index(IndexEntry... children) {
        this(Arrays.asList(children));
    }

    public Index(Collection<IndexEntry> children) {
        this.children = children == null || children.isEmpty()
                ? Collections.<IndexEntry>emptyList()
                : Collections.unmodifiableList(new ArrayList<IndexEntry>(children));
    }

    public static Index read(InputStream inputStream) throws XMLStreamException {
        try {
            return read(XMLInputFactory.newInstance().createXMLStreamReader(inputStream));
        } finally {
            IOUtils.closeQuietly(inputStream);
        }
    }

    public static Index read(XMLStreamReader reader) throws XMLStreamException {
        while (reader.hasNext() && !reader.isStartElement()) {
            reader.next();
        }
        if (reader.getEventType() != XMLStreamConstants.START_ELEMENT) {
            throw new IllegalStateException("Expecting a start element");
        }
        if (!"index".equals(reader.getLocalName())) {
            throw new IllegalStateException("Expecting a <index> element, found a <" + reader.getLocalName() + ">");
        }
        List<IndexEntry> entries = new ArrayList<IndexEntry>();
        int depth = 0;
        while (reader.hasNext() && depth >= 0) {
            switch (reader.next()) {
                case XMLStreamConstants.START_ELEMENT:
                    if (depth == 0 && "entry".equals(reader.getLocalName())) {
                        entries.add(IndexEntry.read(reader));
                    } else {
                        depth++;
                    }
                    break;
                case XMLStreamConstants.END_ELEMENT:
                    depth--;
                    break;
            }
        }
        return new Index(entries);
    }

    public void write(XMLStreamWriter writer) throws XMLStreamException {
        writer.writeStartElement("index");
        for (IndexEntry topic: getChildren()) {
           topic.write(writer);
        }
        writer.writeEndElement();
    }

    public List<IndexEntry> getChildren() {
        return children == null ? Collections.<IndexEntry>emptyList() : children;
    }

    @Override
    public String toString() {
        final StringBuilder sb = new StringBuilder();
        sb.append("Index");
        sb.append("{children=").append(children);
        sb.append('}');
        return sb.toString();
    }
}
