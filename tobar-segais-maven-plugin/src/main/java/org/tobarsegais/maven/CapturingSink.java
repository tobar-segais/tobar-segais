/*
 * Copyright 2013 Stephen Connolly
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

package org.tobarsegais.maven;

import org.apache.maven.doxia.document.DocumentTOCItem;
import org.apache.maven.doxia.module.xhtml.AbstractXhtmlSink;
import org.apache.maven.doxia.sink.SinkEventAttributes;
import org.apache.maven.doxia.util.HtmlTools;

import java.util.ArrayList;
import java.util.List;
import java.util.Stack;

/**
 * @author Stephen Connolly
 */
class CapturingSink extends AbstractXhtmlSink {
    private boolean inTitle = false;
    private StringBuilder buf = new StringBuilder(128);
    private String title;
    private State state = new State(0);
    private Stack<State> stack = new Stack<State>();

    public CapturingSink() {
    }

    @Override
    public void text(String text) {
        buf.append(text);
    }

    @Override
    public void title() {
        buf.setLength(0);
    }

    @Override
    public void title_() {
        title = buf.toString();
    }

    public String getTitle() {
        return title;
    }

    public List<DocumentTOCItem> getItems() {
        if (state.level == 0) {
            return state.items;
        }
        return stack.lastElement().items;
    }

    @Override
    public void sectionTitle(int level, SinkEventAttributes attributes) {
        buf.setLength(0);
    }

    @Override
    public void sectionTitle_(int level) {
        final DocumentTOCItem item = new DocumentTOCItem();
        item.setName(buf.toString());
        item.setRef("#" + HtmlTools.encodeId(buf.toString()));  // TODO figure out a way to let these get rendered
        buf.setLength(0);
        if (level > state.level) {
            List<DocumentTOCItem> items = new ArrayList<DocumentTOCItem>();
            if (!state.items.isEmpty()) {
                state.items.get(state.items.size() - 1).setItems(items);
            }
            stack.push(state);
            state = new State(level, state.items);
            state.items.add(item);
        } else if (level == state.level) {
            state.items.add(item);
        } else {
            while (!stack.isEmpty() && level < state.level) {
                state = stack.pop();
            }
            if (level == state.level) {
                state.items.add(item);
            }
        }
    }

    @Override
    public void sectionTitle1() {
        sectionTitle(1, null);
    }

    @Override
    public void sectionTitle1_() {
        sectionTitle_(1);
    }

    @Override
    public void sectionTitle2() {
        sectionTitle(2, null);
    }

    @Override
    public void sectionTitle2_() {
        sectionTitle_(2);
    }

    @Override
    public void sectionTitle3() {
        sectionTitle(3, null);
    }

    @Override
    public void sectionTitle3_() {
        sectionTitle_(3);
    }

    @Override
    public void sectionTitle4() {
        sectionTitle(4, null);
    }

    @Override
    public void sectionTitle4_() {
        sectionTitle_(4);
    }

    @Override
    public void sectionTitle5() {
        sectionTitle(5, null);
    }

    @Override
    public void sectionTitle5_() {
        sectionTitle_(5);
    }

    private static class State {
        final int level;
        final List<DocumentTOCItem> items;

        private State(int level) {
            this.level = level;
            this.items = new ArrayList<DocumentTOCItem>();
        }

        private State(int level, List<DocumentTOCItem> items) {
            this.level = level;
            this.items = items;
        }
    }

}
