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
import java.util.HashMap;
import java.util.Iterator;
import java.util.List;
import java.util.Map;
import java.util.SortedMap;
import java.util.Stack;
import java.util.TreeMap;

public class Index implements Serializable {

    private static final long serialVersionUID = 1L;
    private final List<IndexEntry> children;
    private final SortedMap<String, IndexEntry> entries;
    private final Map<IndexEntry, String> ids;

    public Index(IndexEntry... children) {
        this(Arrays.asList(children));
    }

    public Index(Collection<IndexEntry> children) {
        this.children = children == null || children.isEmpty()
                ? Collections.<IndexEntry>emptyList()
                : Collections.unmodifiableList(new ArrayList<IndexEntry>(children));
        TreeMap<String,IndexEntry> entryTreeMap = new TreeMap<String, IndexEntry>();
        for (IndexEntry entry: getChildren()) {
            final IndexEntry existing = entryTreeMap.get(entry.getKeyword());
            if (existing != null) {
                entryTreeMap.put(entry.getKeyword(), IndexEntry.merge(existing, entry));
            } else {
                entryTreeMap.put(entry.getKeyword(), entry);
            }
        }
        this.entries = Collections.unmodifiableSortedMap(entryTreeMap);
        Map<IndexEntry, String> ids = new HashMap<IndexEntry, String>();
        int id = 0;
        Stack<Iterator<IndexEntry>> stack = new Stack<Iterator<IndexEntry>>();
        stack.push(entries.values().iterator());
        while (!stack.isEmpty()) {
            Iterator<IndexEntry> iterator = stack.pop();
            while (iterator.hasNext()) {
                IndexEntry indexEntry = iterator.next();
                stack.push(iterator);
                ids.put(indexEntry, Integer.toHexString(id++));
                if (!indexEntry.getSubEntries().isEmpty()) {
                    stack.push(indexEntry.getSubEntries().values().iterator());
                }
            }
        }
        this.ids = Collections.unmodifiableMap(ids);
    }

    public static Index read(String bundle, InputStream inputStream) throws XMLStreamException {
        try {
            return read(bundle, XMLInputFactory.newInstance().createXMLStreamReader(inputStream));
        } finally {
            IOUtils.closeQuietly(inputStream);
        }
    }

    public static Index read(String bundle, XMLStreamReader reader) throws XMLStreamException {
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
                        entries.add(IndexEntry.read(bundle, Collections.<String>emptyList(), reader));
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

    @Override
    public String toString() {
        final StringBuilder sb = new StringBuilder();
        sb.append("Index");
        sb.append("{children=").append(getChildren());
        sb.append('}');
        return sb.toString();
    }

    public List<IndexEntry> getChildren() {
        return children == null ? Collections.<IndexEntry>emptyList() : children;
    }

    public SortedMap<String, IndexEntry> getEntries() {
        return entries;
    }

    public boolean hasChildren() {
        return children != null && !children.isEmpty();
    }

    public IndexEntry findEntry(List<String> keywordPath) {
        return findEntry(keywordPath.iterator());
    }

    private IndexEntry findEntry(Iterator<String> iterator) {
        if (iterator.hasNext()) {
            String keyword = iterator.next();
            IndexEntry indexEntry = getEntries().get(keyword);
            while (indexEntry != null && iterator.hasNext()) {
                keyword = iterator.next();
                indexEntry = indexEntry.getSubEntries().get(keyword);
            }
            return indexEntry;
        }
        return null;
    }

    public String getId(IndexEntry indexEntry) {
        return ids.get(indexEntry);
    }

}
