/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import 'package:flutter/material.dart';
import 'dart:html' as html;

import 'package:playground/config/theme.dart';

enum HeaderAction {
  shortcuts,
  beamPlaygroundGithub,
  apacheBeamGithub,
  scioGithub,
  beamWebsite,
  aboutBeam,
}

const kShortcutsText = "Shortcuts";

const kBeamPlaygroundGithubText = "Beam Playground on GitHub";
const kBeamPlaygroundGithubLink =
    "https://github.com/apache/beam/tree/master/playground/frontend";

const kApacheBeamGithubText = "Apache Beam on GitHub";
const kApacheBeamGithubLink =
    "https://github.com/apache/beam/tree/master/playground/frontend";

const kScioGithubText = "SCIO on GitHub";
const kScioGithubLink = "https://github.com/spotify/scio";

const kBeamWebsiteText = "To Apache Beam website";
const kBeamWebsiteLink = "https://beam.apache.org/";

const kAboutBeamText = "About Apache Beam";

class MoreActions extends StatelessWidget {
  const MoreActions({Key? key}) : super(key: key);

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 16.0, vertical: 8.0),
      child: PopupMenuButton<HeaderAction>(
        icon: Icon(
          Icons.more_horiz_outlined,
          color: ThemeColors.of(context).grey1Color,
        ),
        itemBuilder: (BuildContext context) => <PopupMenuEntry<HeaderAction>>[
          const PopupMenuItem<HeaderAction>(
            padding: EdgeInsets.zero,
            value: HeaderAction.shortcuts,
            child: ListTile(
              leading: Icon(Icons.shortcut_outlined),
              title: Text(kShortcutsText),
            ),
          ),
          PopupMenuItem<HeaderAction>(
            padding: EdgeInsets.zero,
            value: HeaderAction.beamPlaygroundGithub,
            child: ListTile(
              leading: const Icon(Icons.link_outlined),
              title: const Text(kBeamPlaygroundGithubText),
              onTap: () => html.window.open(
                kBeamPlaygroundGithubLink,
                '_blank',
              ),
            ),
          ),
          PopupMenuItem<HeaderAction>(
            padding: EdgeInsets.zero,
            value: HeaderAction.apacheBeamGithub,
            child: ListTile(
              leading: const Icon(Icons.link_outlined),
              title: const Text(kApacheBeamGithubText),
              onTap: () => html.window.open(
                kApacheBeamGithubLink,
                '_blank',
              ),
            ),
          ),
          PopupMenuItem<HeaderAction>(
            padding: EdgeInsets.zero,
            value: HeaderAction.scioGithub,
            child: ListTile(
              leading: const Icon(Icons.link_outlined),
              title: const Text(kScioGithubText),
              onTap: () => html.window.open(
                kScioGithubLink,
                '_blank',
              ),
            ),
          ),
          const PopupMenuDivider(height: 16.0),
          PopupMenuItem<HeaderAction>(
            padding: EdgeInsets.zero,
            value: HeaderAction.beamWebsite,
            child: ListTile(
              leading: const Icon(Icons.link_outlined),
              title: const Text(kBeamWebsiteText),
              onTap: () => html.window.open(
                kBeamWebsiteLink,
                '_blank',
              ),
            ),
          ),
          const PopupMenuItem<HeaderAction>(
            padding: EdgeInsets.zero,
            value: HeaderAction.beamWebsite,
            child: ListTile(
              leading: Icon(Icons.info_outline),
              title: Text(kAboutBeamText),
            ),
          ),
        ],
      ),
    );
  }
}
