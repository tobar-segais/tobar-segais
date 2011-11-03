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
import java.util.Collection;
import java.util.HashMap;
import java.util.Map;

public class Extension {

    private static final long serialVersionUID = 1L;

    private final String point;
    private final Map<String, String> files;

    public Extension(String point, Map<String, String> files) {
        this.point = point;
        this.files = files;
    }

    public static Extension read(XMLStreamReader reader) throws XMLStreamException {
        if (reader.getEventType() != XMLStreamConstants.START_ELEMENT) {
            throw new IllegalStateException("Expecting a start element");
        }
        if (!"extension".equals(reader.getLocalName())) {
            throw new IllegalStateException("Expecting a <extension> element");
        }
        String point = reader.getAttributeValue(null, "point");
        Map<String, String> files = new HashMap<String, String>();
        int depth = 0;
        while (reader.hasNext() && depth >= 0) {
            switch (reader.next()) {
                case XMLStreamConstants.START_ELEMENT:
                    if (depth == 0) {
                        String file = reader.getAttributeValue(null, "file");
                        if (file != null) {
                            files.put(reader.getLocalName(), file);
                        }
                    }
                    depth++;
                    break;
                case XMLStreamConstants.END_ELEMENT:
                    depth--;
                    break;
            }
        }
        return new Extension(point, files);
    }

    public String getPoint() {
        return point;
    }

    public Map<String, String> getFiles() {
        return files;
    }

    public String getFile(String key) {
        return files.get(key);
    }

    public void write(XMLStreamWriter writer) throws XMLStreamException {
        writer.writeStartElement("extension");
        writer.writeAttribute("point", getPoint());
        for (Map.Entry<String, String> file : getFiles().entrySet()) {
            writer.writeStartElement(file.getKey());
            writer.writeAttribute("file", file.getValue());
            writer.writeEndElement();
        }
        writer.writeEndElement();
    }

    @Override
    public String toString() {
        final StringBuilder sb = new StringBuilder();
        sb.append("Extension");
        sb.append("{point='").append(point).append('\'');
        sb.append(", files=").append(files);
        sb.append('}');
        return sb.toString();
    }

    public static Map<String,Extension> toMap(Collection<Extension> extensions) {
        Map<String,Extension> result = new HashMap<String, Extension>(extensions.size());
        for (Extension e: extensions) {
            result.put(e.getPoint(), e);
        }
        return result;
    }
}
