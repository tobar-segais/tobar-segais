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
import java.util.Collections;
import java.util.Iterator;
import java.util.List;

public class IndexSee implements IndexChild {

    private static final long serialVersionUID = 1L;

    private final List<String> keywordPath;

    public IndexSee(String... keywordPath) {
        this(Arrays.asList(keywordPath));
    }
    public IndexSee(List<String> keywordPath) {
        this.keywordPath = Collections.unmodifiableList(keywordPath == null ? Collections.<String>emptyList() : keywordPath);
    }

    public static IndexSee read(XMLStreamReader reader) throws XMLStreamException {
        if (reader.getEventType() != XMLStreamConstants.START_ELEMENT) {
            throw new IllegalStateException("Expecting a start element");
        }
        if (!"see".equals(reader.getLocalName())) {
            throw new IllegalStateException("Expecting a <see> element, got a <" + reader.getLocalName() + ">");
        }
        List<String> keywordPath = new ArrayList<String>();
        keywordPath.add(reader.getAttributeValue(null, "keyword"));
        int depth = 0;
        while (reader.hasNext() && depth >= 0) {
            switch (reader.next()) {
                case XMLStreamConstants.START_ELEMENT:
                    if (depth == 0 && "subpath".equals(reader.getLocalName())) {
                        keywordPath.add(IndexSubpath.read(reader).getKeyword());
                    } else {
                        depth++;
                    }
                    break;
                case XMLStreamConstants.END_ELEMENT:
                    depth--;
                    break;
            }
        }
        return new IndexSee(keywordPath);
    }

    public void write(XMLStreamWriter writer) throws XMLStreamException {
        writer.writeStartElement("see");
        Iterator<String> keywordIterator = keywordPath.iterator();
        if (keywordIterator.hasNext()) {
        writer.writeAttribute("keyword", keywordIterator.next());
        }
        while (keywordIterator.hasNext()) {
            new IndexSubpath(keywordIterator.next()).write(writer);
        }
        writer.writeEndElement();
    }

    public List<String> getKeywordPath() {
        return keywordPath;
    }

    @Override
    public String toString() {
        final StringBuilder sb = new StringBuilder();
        sb.append("IndexSee");
        sb.append("{keywordPath='").append(getKeywordPath()).append('\'');
        sb.append('}');
        return sb.toString();
    }

}
