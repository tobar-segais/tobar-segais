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

import java.io.Serializable;
import java.util.ArrayList;
import java.util.Collection;
import java.util.Collections;
import java.util.List;

public class TocEntry implements Serializable {

    private static final long serialVersionUID = 1L;

    private final String label;
    private final List<Topic> children;
    private final String href;

    public TocEntry(String label, String href, Collection<Topic> children) {
        this.label = label;
        this.children = children == null || children.isEmpty()
                ? Collections.<Topic>emptyList()
                : Collections.unmodifiableList(new ArrayList<Topic>(children));
        this.href = href;
    }

    public String getLabel() {
        return label;
    }

    public List<Topic> getChildren() {
        return children == null ? Collections.<Topic>emptyList() : children;
    }

    public String getHref() {
        return href;
    }

    @Override
    public String toString() {
        final StringBuilder sb = new StringBuilder();
        sb.append("TocEntry");
        sb.append("{label='").append(getLabel()).append('\'');
        sb.append(", href='").append(getHref()).append('\'');
        sb.append(", children=").append(getChildren());
        sb.append('}');
        return sb.toString();
    }

    public Topic lookupTopic(String href) {
        for (Topic topic: getChildren()) {
            if (href.equals(topic.getHref())) return topic;
            Topic r = topic.lookupTopic(href);
            if (r != null) return r;
        }
        return null;
    }
}
