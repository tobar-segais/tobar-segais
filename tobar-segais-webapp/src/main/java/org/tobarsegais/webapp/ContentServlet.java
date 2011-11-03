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

package org.tobarsegais.webapp;

import org.apache.commons.io.IOUtils;

import javax.servlet.ServletException;
import javax.servlet.http.HttpServlet;
import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpServletResponse;
import java.io.IOException;
import java.io.InputStream;
import java.io.OutputStream;
import java.net.JarURLConnection;
import java.net.URL;
import java.net.URLConnection;
import java.util.Map;
import java.util.jar.JarEntry;
import java.util.jar.JarFile;

public class ContentServlet extends HttpServlet {
    @Override
    protected void doGet(HttpServletRequest req, HttpServletResponse resp) throws ServletException, IOException {
        String path = req.getPathInfo();
        if (path == null) {
            path = req.getServletPath();
        }
        int index = path.indexOf("/PLUGINS_ROOT/");
        if (index != -1) {
            path = path.substring(index + "/PLUGINS_ROOT".length());
        }
        Map<String,String> bundles = (Map<String, String>) getServletContext().getAttribute("bundles");
        for (index = path.indexOf('/'); index != -1; index = path.indexOf('/', index + 1)) {
            String key = path.substring(0, index);
            if (key.startsWith("/")) key = key.substring(1);
            if (bundles.containsKey(key)) {
                key = bundles.get(key);
            }
            URL resource = getServletContext().getResource("/WEB-INF/bundles/" + key + ".jar");
            if (resource == null) {
                continue;
            }
            URL jarResource = new URL("jar:" + resource + "!/");
            URLConnection connection = jarResource.openConnection();
            if (!(connection instanceof JarURLConnection)) {
                continue;
            }
            JarURLConnection jarConnection = (JarURLConnection) connection;
            JarFile jarFile = jarConnection.getJarFile();

            int endOfFileName = path.indexOf('#', index);
            endOfFileName = endOfFileName == -1 ? path.length() : endOfFileName;
            String fileName = path.substring(index+1, endOfFileName);
            JarEntry jarEntry = jarFile.getJarEntry(fileName);
            if (jarEntry == null) {
                continue;
            }
            InputStream in = null;
            OutputStream out = resp.getOutputStream();
            try {
                in = jarFile.getInputStream(jarEntry);
                IOUtils.copy(in, out);
            } finally {
                IOUtils.closeQuietly(in);
                out.close();
            }
            return;
        }
        resp.sendError(404);
    }
}
