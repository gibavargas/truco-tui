use adw::subclass::prelude::*;
use adw::gio;
use gtk::glib;

mod imp {
    use super::*;

    #[derive(Debug, Default, gtk::CompositeTemplate)]
    #[template(file = "../window.ui")]
    pub struct TrucoWindow {
        #[template_child]
        pub main_stack: TemplateChild<gtk::Stack>,
        #[template_child]
        pub lobby_page: TemplateChild<gtk::Box>,
        #[template_child]
        pub btn_start_demo: TemplateChild<gtk::Button>,
        #[template_child]
        pub entry_player_name: TemplateChild<gtk::Entry>,
        #[template_child]
        pub dd_num_players: TemplateChild<gtk::DropDown>,
        #[template_child]
        pub dd_locale: TemplateChild<gtk::DropDown>,
        #[template_child]
        pub entry_relay_url: TemplateChild<gtk::Entry>,
        #[template_child]
        pub dd_desired_role: TemplateChild<gtk::DropDown>,
        #[template_child]
        pub btn_host_online: TemplateChild<gtk::Button>,
        #[template_child]
        pub entry_invite_key: TemplateChild<gtk::Entry>,
        #[template_child]
        pub btn_join_online: TemplateChild<gtk::Button>,
        #[template_child]
        pub lbl_online_status: TemplateChild<gtk::Label>,
        #[template_child]
        pub lbl_invite_key_display: TemplateChild<gtk::Label>,
        #[template_child]
        pub list_slots: TemplateChild<gtk::ListBox>,
        #[template_child]
        pub btn_start_online_match: TemplateChild<gtk::Button>,
        #[template_child]
        pub list_chat: TemplateChild<gtk::ListBox>,
        #[template_child]
        pub entry_chat: TemplateChild<gtk::Entry>,
        #[template_child]
        pub btn_send_chat: TemplateChild<gtk::Button>,
        #[template_child]
        pub btn_leave_online: TemplateChild<gtk::Button>,
        
        #[template_child]
        pub game_page: TemplateChild<gtk::Overlay>,
        #[template_child]
        pub hud_box: TemplateChild<gtk::Box>,
        #[template_child]
        pub opponent_box: TemplateChild<gtk::Box>,
        #[template_child]
        pub left_player_box: TemplateChild<gtk::Box>,
        #[template_child]
        pub right_player_box: TemplateChild<gtk::Box>,
        #[template_child]
        pub center_box: TemplateChild<gtk::Box>,
        #[template_child]
        pub bottom_box: TemplateChild<gtk::Box>,
        #[template_child]
        pub game_over_overlay: TemplateChild<gtk::Overlay>,
        #[template_child]
        pub lbl_winner: TemplateChild<gtk::Label>,
        #[template_child]
        pub btn_back_lobby: TemplateChild<gtk::Button>,
        #[template_child]
        pub log_box: TemplateChild<gtk::Box>,
        #[template_child]
        pub btn_leave_match: TemplateChild<gtk::Button>,
        #[template_child]
        pub trick_end_overlay: TemplateChild<gtk::Overlay>,
        #[template_child]
        pub lbl_trick_emoji: TemplateChild<gtk::Label>,
        #[template_child]
        pub lbl_trick_text: TemplateChild<gtk::Label>,
    }

    #[glib::object_subclass]
    impl ObjectSubclass for TrucoWindow {
        const NAME: &'static str = "TrucoWindow";
        type Type = super::TrucoWindow;
        type ParentType = adw::ApplicationWindow;

        fn class_init(klass: &mut Self::Class) {
            klass.bind_template();
        }

        fn instance_init(obj: &glib::subclass::InitializingObject<Self>) {
            obj.init_template();
        }
    }

    impl ObjectImpl for TrucoWindow {}
    impl WidgetImpl for TrucoWindow {}
    impl WindowImpl for TrucoWindow {}
    impl ApplicationWindowImpl for TrucoWindow {}
    impl adw::subclass::prelude::AdwApplicationWindowImpl for TrucoWindow {}
}

glib::wrapper! {
    pub struct TrucoWindow(ObjectSubclass<imp::TrucoWindow>)
        @extends gtk::Widget, gtk::Window, gtk::ApplicationWindow, adw::ApplicationWindow,
        @implements gio::ActionMap, gio::ActionGroup, gtk::Root;
}

impl TrucoWindow {
    pub fn new(app: &adw::Application) -> Self {
        glib::Object::builder().property("application", app).build()
    }

    pub fn main_stack(&self) -> gtk::Stack { self.imp().main_stack.get() }
    pub fn lobby_page(&self) -> gtk::Box { self.imp().lobby_page.get() }
    pub fn btn_start_demo(&self) -> gtk::Button { self.imp().btn_start_demo.get() }
    pub fn entry_player_name(&self) -> gtk::Entry { self.imp().entry_player_name.get() }
    pub fn dd_num_players(&self) -> gtk::DropDown { self.imp().dd_num_players.get() }
    pub fn dd_locale(&self) -> gtk::DropDown { self.imp().dd_locale.get() }
    pub fn entry_relay_url(&self) -> gtk::Entry { self.imp().entry_relay_url.get() }
    pub fn dd_desired_role(&self) -> gtk::DropDown { self.imp().dd_desired_role.get() }
    pub fn btn_host_online(&self) -> gtk::Button { self.imp().btn_host_online.get() }
    pub fn entry_invite_key(&self) -> gtk::Entry { self.imp().entry_invite_key.get() }
    pub fn btn_join_online(&self) -> gtk::Button { self.imp().btn_join_online.get() }
    pub fn lbl_online_status(&self) -> gtk::Label { self.imp().lbl_online_status.get() }
    pub fn lbl_invite_key_display(&self) -> gtk::Label { self.imp().lbl_invite_key_display.get() }
    pub fn list_slots(&self) -> gtk::ListBox { self.imp().list_slots.get() }
    pub fn btn_start_online_match(&self) -> gtk::Button { self.imp().btn_start_online_match.get() }
    pub fn list_chat(&self) -> gtk::ListBox { self.imp().list_chat.get() }
    pub fn entry_chat(&self) -> gtk::Entry { self.imp().entry_chat.get() }
    pub fn btn_send_chat(&self) -> gtk::Button { self.imp().btn_send_chat.get() }
    pub fn btn_leave_online(&self) -> gtk::Button { self.imp().btn_leave_online.get() }

    pub fn game_page(&self) -> gtk::Overlay { self.imp().game_page.get() }
    pub fn hud_box(&self) -> gtk::Box { self.imp().hud_box.get() }
    pub fn opponent_box(&self) -> gtk::Box { self.imp().opponent_box.get() }
    pub fn left_player_box(&self) -> gtk::Box { self.imp().left_player_box.get() }
    pub fn right_player_box(&self) -> gtk::Box { self.imp().right_player_box.get() }
    pub fn center_box(&self) -> gtk::Box { self.imp().center_box.get() }
    pub fn bottom_box(&self) -> gtk::Box { self.imp().bottom_box.get() }
    pub fn game_over_overlay(&self) -> gtk::Overlay { self.imp().game_over_overlay.get() }
    pub fn lbl_winner(&self) -> gtk::Label { self.imp().lbl_winner.get() }
    pub fn btn_back_lobby(&self) -> gtk::Button { self.imp().btn_back_lobby.get() }
    pub fn log_box(&self) -> gtk::Box { self.imp().log_box.get() }
    pub fn btn_leave_match(&self) -> gtk::Button { self.imp().btn_leave_match.get() }
    pub fn trick_end_overlay(&self) -> gtk::Overlay { self.imp().trick_end_overlay.get() }
    pub fn lbl_trick_emoji(&self) -> gtk::Label { self.imp().lbl_trick_emoji.get() }
    pub fn lbl_trick_text(&self) -> gtk::Label { self.imp().lbl_trick_text.get() }
}
