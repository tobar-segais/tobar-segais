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
import java.util.ArrayList;
import java.util.Arrays;
import java.util.Collection;
import java.util.List;
import java.util.Map;

public class Plugin {

    private static final long serialVersionUID = 1L;

    private final String name;
    private final String id;
    private final String version;
    private final String providerName;
    private final Map<String,Extension> extensions;

    public Plugin(String name, String id, String version, String providerName, Extension... extensions) {
        this(name, id, version, providerName, Extension.toMap(Arrays.asList(extensions)));
    }

    public Plugin(String name, String id, String version, String providerName, Map<String, Extension> extensions) {
        this.name = name;
        this.id = id;
        this.version = version;
        this.providerName = providerName;
        this.extensions = extensions;
    }

    public Plugin(String name, String id, String version, String providerName,  Collection<Extension> extensions) {
        this(name, id, version, providerName, Extension.toMap(extensions));
    }

    public static Plugin read(InputStream inputStream) throws XMLStreamException {
        try {
            return read(XMLInputFactory.newInstance().createXMLStreamReader(inputStream));
        } finally {
            IOUtils.closeQuietly(inputStream);
        }
    }

    public static Plugin read(XMLStreamReader reader) throws XMLStreamException {
        while (reader.hasNext() && !reader.isStartElement()) {
            reader.next();
        }
        if (reader.getEventType() != XMLStreamConstants.START_ELEMENT) {
            throw new IllegalStateException("Expecting a start element");
        }
        if (!"plugin".equals(reader.getLocalName())) {
            throw new IllegalStateException("Expecting a <plugin> element");
        }
        String name = reader.getAttributeValue(null, "name");
        String id = reader.getAttributeValue(null, "id");
        String version = reader.getAttributeValue(null, "version");
        String providerName = reader.getAttributeValue(null, "provider-name");
        List<Extension> extensions = new ArrayList<Extension>();
        int depth = 0;
        while (reader.hasNext() && depth >= 0) {
            switch (reader.next()) {
                case XMLStreamConstants.START_ELEMENT:
                    if (depth == 0 && "extension".equals(reader.getLocalName())) {
                        extensions.add(Extension.read(reader));
                    } else {
                        depth++;
                    }
                    break;
                case XMLStreamConstants.END_ELEMENT:
                    depth--;
                    break;
            }
        }
        return new Plugin(name, id, version, providerName, extensions);
    }

    public String getName() {
        return name;
    }

    public String getId() {
        return id;
    }

    public String getVersion() {
        return version;
    }

    public String getProviderName() {
        return providerName;
    }

    public Map<String, Extension> getExtensions() {
        return extensions;
    }

    public Extension getExtension(String point) {
        return extensions.get(point);
    }

    public void write(XMLStreamWriter writer) throws XMLStreamException {
        writer.writeStartElement("plugin");
        writer.writeAttribute("name", getName());
        writer.writeAttribute("id", getId());
        writer.writeAttribute("version", getVersion());
        writer.writeAttribute("provider-name", getProviderName());
        for (Extension extension: getExtensions().values()) {
           extension.write(writer);
        }
        writer.writeEndElement();
    }

}
