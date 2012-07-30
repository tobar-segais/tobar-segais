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
import java.util.SortedMap;
import java.util.TreeMap;

public class IndexEntry implements IndexChild {

    private static final long serialVersionUID = 1L;

    private final List<String> path;
    private final String keyword;

    private final List<IndexTopic> topics;
    private final List<IndexSee> sees;
    private final SortedMap<String,IndexEntry> subEntries;

    public IndexEntry(List<String> path,String keyword, Collection<IndexTopic> topics, Collection<IndexSee> sees, Collection<IndexEntry> subEntries) {
        this.path = path == null ? Collections.<String>emptyList() : path;
        this.keyword = keyword;
        this.topics = topics == null ? Collections.<IndexTopic>emptyList() : Collections.unmodifiableList(
                new ArrayList<IndexTopic>(topics));
        this.sees = sees == null ? Collections.<IndexSee>emptyList() : Collections.unmodifiableList(new ArrayList
                <IndexSee>(sees));
        TreeMap<String,IndexEntry> entryTreeMap = new TreeMap<String, IndexEntry>();
        for (IndexEntry entry: subEntries) {
            final IndexEntry existing = entryTreeMap.get(entry.getKeyword());
            if (existing != null) {
                entryTreeMap.put(entry.getKeyword(), merge(existing, entry));
            } else {
                entryTreeMap.put(entry.getKeyword(), entry);
            }
        }
        this.subEntries = Collections.unmodifiableSortedMap(entryTreeMap);
    }

    public static IndexEntry merge(IndexEntry entry, IndexEntry... entries) {
        final List<IndexTopic> topics = new ArrayList<IndexTopic>(entry.getTopics());
        final List<IndexSee> sees = new ArrayList<IndexSee>(entry.getSees());
        final List<IndexEntry> subEntries = new ArrayList<IndexEntry>(entry.getSubEntries().values());
        for (IndexEntry e: entries) {
            topics.addAll(e.getTopics());
            sees.addAll(e.getSees());
            subEntries.addAll(e.getSubEntries().values());
        }
        return new IndexEntry(entry.getPath(), entry.getKeyword(), topics, sees, subEntries);
    }

    public static IndexEntry read(String bundle, List<String> path, XMLStreamReader reader) throws XMLStreamException {
        if (reader.getEventType() != XMLStreamConstants.START_ELEMENT) {
            throw new IllegalStateException("Expecting a start element");
        }
        if (!"entry".equals(reader.getLocalName())) {
            throw new IllegalStateException("Expecting a <entry> element, got a <" + reader.getLocalName() + ">");
        }
        String keyword = reader.getAttributeValue(null, "keyword");
        final List<IndexTopic> topics = new ArrayList<IndexTopic>();
        final List<IndexSee> sees = new ArrayList<IndexSee>();
        final List<IndexEntry> subEntries = new ArrayList<IndexEntry>();
        List<String> childPath = new ArrayList<String>(path);
        childPath.add(keyword);
        int depth = 0;
        while (reader.hasNext() && depth >= 0) {
            switch (reader.next()) {
                case XMLStreamConstants.START_ELEMENT:
                    if (depth == 0 && "topic".equals(reader.getLocalName())) {
                        topics.add(IndexTopic.read(bundle, reader));
                    } else if (depth == 0 && "entry".equals(reader.getLocalName())) {
                        subEntries.add(IndexEntry.read(bundle, childPath, reader));
                    } else if (depth == 0 && "see".equals(reader.getLocalName())) {
                        sees.add(IndexSee.read(reader));
                    } else {
                        depth++;
                    }
                    break;
                case XMLStreamConstants.END_ELEMENT:
                    depth--;
                    break;
            }
        }
        return new IndexEntry(path, keyword, topics, sees, subEntries);
    }

    public void write(XMLStreamWriter writer) throws XMLStreamException {
        writer.writeStartElement("entry");
        writer.writeAttribute("keyword", getKeyword());
        for (IndexTopic topic : getTopics()) {
            topic.write(writer);
        }
        for (IndexSee see : getSees()) {
            see.write(writer);
        }
        for (IndexEntry entry : getSubEntries().values()) {
            entry.write(writer);
        }
        writer.writeEndElement();
    }

    public String getKeyword() {
        return keyword;
    }

    @Override
    public String toString() {
        final StringBuilder sb = new StringBuilder();
        sb.append("IndexEntry");
        sb.append("{keyword='").append(getKeyword()).append('\'');
        sb.append(", topics=").append(getTopics());
        sb.append(", sees=").append(getSees());
        sb.append(", subEntries=").append(getSubEntries());
        sb.append('}');
        return sb.toString();
    }

    public List<String> getPath() {
        return path;
    }

    public List<IndexTopic> getTopics() {
        return topics;
    }

    public List<IndexSee> getSees() {
        return sees;
    }

    public SortedMap<String, IndexEntry> getSubEntries() {
        return subEntries;
    }

    public boolean hasChildren() {
        return !getTopics().isEmpty() || !getSees().isEmpty() || !getSubEntries().isEmpty();
    }
}
